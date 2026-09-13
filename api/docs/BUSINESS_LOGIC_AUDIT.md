# Audit métier de l'API httpSMS

Documentation complète du fonctionnement métier de l'API (`api/`) et liste des incohérences détectées, produites par relecture intégrale du code (handlers, services, repositories, entities, middlewares, listeners) le 2026-09-09.

Périmètre couvert : Auth & Users, Phones/Phone API Keys/Heartbeats, Messages (v1/v2/threads/bulk), Billing/Apps/LemonSqueezy, Webhooks/Events/Intégrations (Discord, 3CX).

---

## Résumé exécutif — à traiter en priorité

### 🔴 Critique (crash serveur, perte financière, faille de sécurité, fonctionnalité core cassée)

1. **Aucun middleware `recover` global** — `fiber.New()` sans garde-fou. Le moindre panic non rattrapé (ex: nil-pointer deref, slice out of range) fait tomber **tout le processus pour tous les utilisateurs**, pas juste la requête fautive. C'est la cause du crash déjà rencontré et corrigé cette session (`bearer_auth_middleware.go`). *(Domaine Auth & Users, §Incohérences #2)*
2. **2 nil-pointer dereferences réels et atteignables** sur les endpoints d'abonnement (`DELETE /v1/users/subscription`, `GET /v1/users/subscription-update-url`) — tout utilisateur en plan gratuit (majorité des comptes) qui appelle ces routes crashe le serveur entier, combiné au point précédent. *(Auth & Users #3)*
3. **Plans annuels 50k/100k/200k traités comme "free"** — un client qui paie un de ces plans annuels est silencieusement classé `free` côté facturation/limites (aucune constante définie, le mapping LemonSqueezy ne matche que sur "monthly"). Impact financier direct. *(Billing & Apps #1)*
4. **`phone_notification_listener.go` n'est jamais enregistré** — si rien d'autre ne duplique ce rôle, les notifications FCM déclenchant l'envoi réel de SMS par le téléphone (`message.api.sent`, `message.send.retry`) et les alertes heartbeat manqué ne partent jamais. **À vérifier en urgence** : est-ce que l'envoi de SMS fonctionne réellement en prod par un autre chemin ? *(Webhooks & Events #1)*
5. **Auth par clé App (`ak_`) totalement non fonctionnelle** — `AppAuth` ne peuple jamais `Email` dans le contexte d'auth, donc `IsNoop()` le traite toujours comme non-authentifié. Les endpoints `/v1/apps/messages/send` et `/status` sont inutilisables avec une vraie clé `ak_`/`x-api-secret`. *(Auth & Users #1)*
6. **v2 `POST /v2/send` sans vérification de possession du numéro expéditeur** — n'importe quel titulaire d'une clé `uk_` valide peut prétendre envoyer depuis n'importe quel numéro syntaxiquement valide, sans vérifier qu'il lui appartient (contrairement à tous les endpoints v1 protégés par `authorizePhoneAPIKey`). *(Messages #7)*
7. **Bug d'inversion de condition — le token FCM est effacé silencieusement** à chaque modification d'un téléphone via le dashboard web (`PUT /v1/phones`), cassant les notifications push pour ce téléphone. *(Phones & Heartbeats #2)*
8. **Aucune validation au démarrage de la config de la file d'événements** (`EVENTS_QUEUE_USER_API_KEY`/`_ID`) — une mauvaise config fait échouer TOUT le pipeline d'événements silencieusement (vécu réellement cette session : heartbeats/facturation/notifications bloqués pendant des heures sans erreur visible). *(Webhooks & Events #2, déjà expérimenté)*

### 🟠 Majeur (bug fonctionnel silencieux, incohérence notable)

- Asymétrie `PUT /v1/phones` vs `/fcm-token` : un numéro ajouté via le dashboard n'est jamais lié à une clé API téléphone, mais un `HeartbeatMonitor` est quand même créé pour lui → heartbeats voués à échouer indéfiniment tant que l'app Android n'a pas fait son propre login. *(Phones #3 — c'est exactement l'incident réel de cette session)*
- Gestion d'erreur incohérente entre fonctions sœurs du domaine Messages (`HandleMessageSent`/`HandleMessageDelivered`/`HandleMessageFailed` ne traitent pas les statuts inattendus/doublons de la même façon). *(Messages #1, #2)*
- Duplication complète de la logique d'envoi en masse entre `BulkSend` (JSON) et `BulkMessageHandler.Store` (CSV) — tout bug de concurrence/limite doit être corrigé deux fois. *(Messages #13)*
- `emulatorPushQueue` (dev) n'a ni retry ni remontée d'erreur, contrairement à la queue Cloud Tasks (prod) — un event perdu en dev est indétectable sans grep manuel des logs. *(Webhooks #3)*
- Deux systèmes de webhooks sortants incompatibles (V1 JWT sans hash du payload / V2 HMAC du payload), avec couverture d'events et retry différents. *(Webhooks #4)*
- `IsEntitledWithCount` "fail open" : une panne DB lève silencieusement TOUTES les limites d'envoi pour tous les utilisateurs, sans alerte distincte. *(Billing #5)*
- `gormBillingUsageRepository` utilise des primitives de retry spécifiques CockroachDB (`crdbgorm`) alors que le dev tourne sur Postgres — à vérifier si la prod est vraiment sur CockroachDB. *(Billing #6)*
- Le logout ne révoque aucun JWT émis (stateless, 24h de validité résiduelle après déconnexion). *(Auth #5)*
- `AddPhone` (association manuelle téléphone↔clé) n'a aucune route HTTP — impossible de rattraper une association manquante sans repasser par le flux complet ou une panne d'events. *(Phones #4)*

### 🟡 Mineur (cosmétique, dette technique, incohérence de nommage/doc)

Fautes de frappe (`middlesare`, "thread thread", "TIMOUT"), messages d'erreur incohérents avec la valeur réelle (limite webhooks 10 vs message disant "5" dans les logs), conventions de nommage d'events incohérentes (`EventType*` vs sans préfixe), code mort documenté comme tel (`DeleteAuthUser`, 4/5 listeners 3CX désactivés), `spew.Dump` inconditionnel du corps des requêtes 3CX en logs (fuite potentielle de contenu SMS) — détail complet dans les sections par domaine ci-dessous.

---

## 1. Domaine : Auth & Users

### Documentation

#### Architecture d'authentification

Le serveur supporte **5 sources d'authentification simultanées**, toutes évaluées par des middlewares globaux ou semi-globaux qui peuplent `c.Locals(ContextKeyAuthUserID)` avec un `entities.AuthContext{ID, Email, PhoneAPIKeyID, AppID, PhoneNumbers}` :

| Source | Constante | Header | Middleware | Enregistrement |
|---|---|---|---|---|
| Cookie de session | `AuthSourceSession` | Cookie `httpsms_session` | `SessionAuth` | Global (`app.Use`, container.go:198) |
| JWT Bearer | `AuthSourceBearer` | `Authorization: Bearer <jwt>` | `BearerAuth` | Global (container.go:199) |
| Clé API utilisateur | `AuthSourceAPIKey` | `x-api-key: uk_...` (tout sauf `pk_`) | `APIKeyAuth` | Global (container.go:200) |
| Clé API téléphone | `AuthSourcePhoneKey` | `x-api-key: pk_...` | `PhoneAPIKeyAuth` | Par-route via `PhoneAPIKeyMiddleware()`, jamais globale |
| Clé API app + secret | `AuthSourceAppKey` | `x-api-key: ak_...` + `x-api-secret` | `AppAuth` | Par-route via `AppAuthMiddleware()` (container.go:1609) |
| Bearer API key (legacy) | `AuthSourceAPIKey` | `Authorization: Bearer <uk_key>` | `BearerAPIKeyAuth` | Par-route via `BearerAPIKeyMiddleware()` |

Chaque middleware, s'il échoue à authentifier, appelle simplement `c.Next()` sans bloquer — l'échec définitif n'intervient qu'au niveau de `Authenticated` (middleware générique, container.go `AuthenticatedMiddleware()`), qui vérifie `!tokenUser.IsNoop()`.

`entities.AuthContext.IsNoop()` (auth_context.go:15-17) :
```go
return user.ID == "" || user.Email == ""
```
**Toute source d'auth qui ne peuple pas `Email` est donc traitée comme non-authentifiée**, quel que soit `ID`.

#### `handler.go` — base commune à tous les handlers

- Fonctions `response*` : enveloppe JSON standard `{status, message, data}`.
- `userFromContext(c)` (ligne 115-120) : récupère l'`AuthContext`, **`panic("user does not exist in context.")`** si absent/noop. Aucun `recover()` n'est enregistré dans `container.go`.
- `authorizePhoneAPIKey(c, phoneNumber)` (ligne 142-148) : si `PhoneAPIKeyID == nil` → autorisé par défaut (seules les requêtes authentifiées via clé téléphone sont réellement restreintes à une liste de numéros).

#### Routes `AuthHandler` (`/v1/auth/*`) — publiques, aucun middleware

- `POST /v1/auth/register` → `AuthService.Register` : unicité email (409 si conflit), hash bcrypt, génère `api_key` (`uk_` + 64 chars aléatoires), crée une session (30 jours par défaut ou `SESSION_DURATION`), pose le cookie, **génère un JWT** (ajouté cette session) et le renvoie avec `user`. Dispatche `user.account.created`.
- `POST /v1/auth/login` → `AuthService.Login` : bcrypt, nouvelle session à chaque login, cookie, JWT, `{user, token}`.
- `DELETE /v1/auth/sessions` (logout) : supprime la session en DB si trouvée, clear le cookie. **Ne révoque aucun JWT émis.**
- Cookie de session (`setSessionCookie`) : `Secure`/`SameSite=None` si `c.Protocol()=="https"` OU `ENV=production` ; sinon `SameSite=Lax`. Corrigé cette session pour le cross-origin (tunnel Cloudflare).

#### Routes `UserHandler` (`/v1/users/*`) — `AuthenticatedMiddleware()` uniquement

| Route | Méthode | Fonction |
|---|---|---|
| `/v1/users/me` | GET | `Show` → `UserService.Get` (crée l'utilisateur s'il n'existe pas — `LoadOrStore`) |
| `/v1/users/me` | PUT | `Update` → timezone, active_phone_id, webhook_url |
| `/v1/users/me` | DELETE | `Delete` → refuse si abonnement payant actif non expiré |
| `/v1/users/:userID/api-keys` | DELETE | `DeleteAPIKey` (rotation) → vérifie `userID` du path == user du contexte |
| `/v1/users/:userID/notifications` | PUT | `UpdateNotifications` — aucune validation |
| `/v1/users/subscription-update-url` | GET | URL du portail client LemonSqueezy |
| `/v1/users/subscription` | DELETE | Annulation LemonSqueezy |
| `/v1/users/subscription/payments` | GET | Historique des paiements |
| `/v1/users/subscription/invoices/:subscriptionInvoiceID` | POST | Génère un PDF de facture |

#### Modèles

- **`entities.User`** : `ID`, `APIKey` (`uk_...`), `SubscriptionName/ID/Status/RenewsAt/EndsAt`, 4 flags de notification, `WebhookURL`, `Timezone` (défaut `Africa/Accra`), `ActivePhoneID`.
- **`entities.Session`** : `ID`, `UserID`, `Email` (dupliqué depuis User à la création), `ExpiresAt`.
- **JWT** (`AuthService.GenerateToken`) : claims `user_id`/`email`/`exp` (24h fixe)/`iat`, signé HS256 (`JWT_SECRET`). Validé par `BearerAuth` (rejette tout ce qui n'est pas HMAC). Pas de révocation.

### Incohérences détectées

1. **[middlewares/app_auth_middleware.go:48-52] Auth par clé App (`ak_`) totalement non fonctionnelle.** `AppAuth` construit `AuthContext{ID, AppID, PhoneNumbers}` sans jamais peupler `Email`. `IsNoop()` exige `Email != ""` → 401 systématique même avec un couple `ak_`/`x-api-secret` valide. `POST /v1/apps/messages/send` et `GET /v1/apps/messages/:messageID/status` sont inutilisables en l'état.

2. **[handler.go:119 + absence de `recover`] Aucun filet contre les panics.** `fiber.New()` sans `middleware/recover`. Le moindre panic crashe tout le process pour tous les utilisateurs.

3. **[services/user_service.go:429-432, 439-455] Nil-pointer deref si un utilisateur gratuit appelle les endpoints d'abonnement.** `InitiateSubscriptionCancel` et `GetSubscriptionUpdateURL` font `*user.SubscriptionID` sans checker `nil`, contrairement à `GetSubscriptionPayments` (ligne 69) qui le fait. Un utilisateur free (`SubscriptionID == nil`) qui appelle `DELETE /v1/users/subscription` ou `GET /v1/users/subscription-update-url` crashe le serveur.

4. **[handlers/user_handler.go:163-182] `UpdateNotifications` ne valide jamais son payload** — seul endpoint mutant sans validateur dédié.

5. **[handlers/auth_handler.go, logout] Le logout ne révoque pas les JWT émis** — valides 24h après déconnexion, aucune blacklist.

6. **[entities/session.go:13] `Session.Email` peut devenir périmé** — copié à la création, jamais resynchronisé si l'utilisateur change d'email.

7. **[services/user_service.go:539-546] `DeleteAuthUser` est un no-op documenté comme tel, mais toujours appelé** — reliquat de la migration Firebase → auth interne, code mort à nettoyer.

8. **[middlewares/authenticated_middlesare.go] Faute de frappe dans le nom du fichier** (`middlesare` au lieu de `middleware`), présente dans le repo.

9. **[handler.go:33-39] Message d'erreur générique trompeur** — `responseUnauthorized` répond toujours "Make sure your API key is set in [X-API-Key]" même pour un cookie expiré, un JWT invalide, ou une clé App mal câblée (#1). A directement compliqué le diagnostic cette session.

---

## 2. Domaine : Phones, Phone API Keys & Heartbeats

### Documentation

#### Modèle de données

- **`entities.Phone`** : `ID`, `UserID`, `PhoneNumber` (E164), `FcmToken` (*string), `SIM`, `MessagesPerMinute`, `MaxSendAttempts`, `MessageExpirationSeconds`, `MissedCallAutoReply`. Pas de champ `WebhookURL`. Pas de lien direct vers une `PhoneAPIKey` — la relation vit uniquement côté `PhoneAPIKey.PhoneNumbers`/`PhoneIDs`.
- **`entities.PhoneAPIKey`** : `PhoneNumbers`/`PhoneIDs` (tableaux Postgres dénormalisés, pas de table de jointure). **Règle implicite : un téléphone n'appartient qu'à une seule clé à la fois** — `AddPhone` retire d'abord le téléphone de toutes les clés de l'utilisateur avant de l'ajouter à la cible.
- **`entities.HeartbeatMonitor`** : état par (UserID, Owner), `PhoneOnline`, `QueueID`, `RequiresCheck()` (non mis à jour depuis 2h), `PhoneIsOffline()`.

#### Routes & middlewares réels

| Route | Handler | Middlewares | Usage |
|---|---|---|---|
| `GET/PUT /v1/phones`, `DELETE /v1/phones/:id` | `PhoneHandler` | `AuthenticatedMiddleware` | Dashboard web |
| `PUT /v1/phones/fcm-token` | `UpsertFCMToken` | `PhoneAPIKeyMiddleware` + `AuthenticatedMiddleware` | App Android uniquement |
| CRUD `/v1/phone-api-keys/*` | `PhoneAPIKeyHandler` | `AuthenticatedMiddleware` | Dashboard — seule opération d'association exposée est un **retrait** (`deletePhone`) |
| `GET /v1/heartbeats` | `Index` | `AuthenticatedMiddleware` | Dashboard |
| `POST /v1/heartbeats` | `Store` | `PhoneAPIKeyMiddleware` + `AuthenticatedMiddleware` | App Android uniquement |

`PhoneAPIKeyService.AddPhone` n'a **aucune route HTTP dédiée** — déclenché uniquement par le listener d'événement.

#### Flux d'association téléphone ↔ clé API

Deux chemins totalement différents pour "mettre à jour" un `Phone` :

1. **`PUT /v1/phones` → `Upsert`** (dashboard, session/JWT/uk_) : construit l'event `PhoneUpdatedPayload` avec **`PhoneAPIKeyID: nil` codé en dur**. `PhoneAPIKeyListener.onPhoneUpdated` voit `nil` et retourne sans appeler `AddPhone` → **aucune association créée**. `HeartbeatListener.onPhoneUpdated`, lui, ne regarde pas `PhoneAPIKeyID` — un monitor est créé quand même (incohérence #3).
2. **`PUT /v1/phones/fcm-token` → `UpsertFCMToken`** (app Android, clé `pk_`) : `PhoneAPIKeyID: user.PhoneAPIKeyID` récupéré du contexte d'auth réel → event dispatché avec le vrai ID → `AddPhone` appelé → `phone_numbers`/`phone_ids` mis à jour. **Seul chemin qui crée l'association.**

#### Autorisation au niveau requête

```go
func (h *handler) authorizePhoneAPIKey(c *fiber.Ctx, phoneNumber string) bool {
    user := h.userFromContext(c)
    if user.PhoneAPIKeyID == nil {
        return true
    }
    return slices.Contains(user.PhoneNumbers, phoneNumber)
}
```
Utilisé par `HeartbeatHandler.Store` et les endpoints de messages entrants/événements. `AuthContext.PhoneNumbers` chargé depuis `phone_api_keys.phone_numbers` avec cache ristretto TTL 15s.

#### Pipeline heartbeat/monitoring

`POST /v1/heartbeats` → sauvegarde par numéro (goroutines) → si monitor existant et `PhoneIsOffline()`, event `phone.heartbeat.online`. Monitor créé/entretenu via `phone.updated` et `phone.heartbeat.check` (auto-replanifié toutes les 16 min). `HeartbeatService.Monitor` : retard 16-80min → `phone.heartbeat.missed` ; retard 64-80min ET online → `phone.heartbeat.offline` (chevauchement, incohérence #5). Dépend entièrement de `/v1/events` fonctionnel.

### Incohérences détectées

1. **[services/phone_service.go:90] Champ mort `WebhookURL` sur `PhoneUpsertParams`** — jamais lu par `update()`, absent de l'entité et de la requête.

2. **[services/phone_service.go:277-279] Bug d'inversion de condition — FCM token effacé silencieusement.**
   ```go
   if phone.FcmToken != nil {       // teste la valeur EXISTANTE
       phone.FcmToken = params.FcmToken
   }
   ```
   Devrait tester `params.FcmToken != nil` (la valeur ENTRANTE). Tout `PUT /v1/phones` depuis le dashboard efface le token FCM existant → notifications push cassées silencieusement à chaque édition de réglage.

3. **Asymétrie dashboard/app sur l'association clé API** (détaillée ci-dessus) — cause racine de l'incident réel de cette session (numéros ajoutés côté dashboard, heartbeats rejetés en 401 malgré un monitor existant).

4. **`AddPhone` n'a pas de route HTTP** — impossible de lier manuellement un téléphone sans repasser par le flux complet ou en cas de panne de la file d'événements (exactement ce qui s'est produit).

5. **[services/heartbeat_service.go:319-327] Fenêtre de chevauchement missed/offline** — entre 64 et 80 min de retard, les deux events (`missed` ET `offline`) partent pour le même monitor, sans exclusion mutuelle. Notification en double si non intentionnel.

6. **[handlers/heartbeat_handler.go:140-156] Échecs silencieux dans `Store`** — chaque numéro traité en goroutine séparée ; une erreur est seulement journalisée, jamais remontée au client (toujours `201 Created`).

7. **Commentaires copié-collés sur `RegisterPhoneAPIKeyRoutes`** (phone_handler.go:47, heartbeat_handler.go:47-48) ne précisent pas qu'il s'agit spécifiquement des routes clé-téléphone — piège de lecture dans un domaine où cette distinction est critique.

---

## 3. Domaine : Messages (v1, v2, threads, bulk)

### Documentation

#### Modèle de données

- **`entities.Message`** : machine à états. Types `mobile-terminated`/`mobile-originated`/`call-missed`. Statuts `pending → scheduled → sending → sent → delivered`, branches `failed`/`expired`/`received`/`deleted`. Méthodes `IsX()` pour Sending/Delivered/Pending/Scheduled/Expired/Sent — **pas** pour Received ni Deleted.
- **`entities.MessageThread`** : vue agrégée par (owner, contact), dernier message/statut, couleur d'affichage aléatoire, `is_archived`.

#### Routes v1 (deux groupes de middlewares distincts)

**`RegisterRoutes`** (clé compte générale `uk_`/session/JWT) : `POST /v1/messages/send`, `POST /v1/messages/bulk-send`, `GET /v1/messages`, `/search`, `/:id`, `DELETE /v1/messages/:id`.

**`RegisterPhoneAPIKeyRoutes`** (clé téléphone `pk_`, appelé par l'app) : `POST /v1/messages/:messageID/events` (ack sent/delivered/failed, vérifie `authorizePhoneAPIKey`), `POST /v1/messages/receive` (MO, vérifie billing + auth), `POST /v1/messages/calls/missed`, `GET /v1/messages/outstanding`.

Séparation cohérente et intentionnelle (envoi = compte, réception/ack = téléphone précis).

#### Route v2 — totalement indépendante

`POST /v2/send` : pas de middleware partagé, auth inline (`x-api-key` uk_ uniquement, pas de session/JWT/pk_), validation inline (sans passer par le validateur standard), réponses `{"error": "CODE"}` au lieu du format standard, pas de `sim`/`SendAt`/`RequestID`.

#### Threads

`GET/PUT/DELETE /v1/message-threads[/:id]`. `UpdateThread` (interne, appelé par listener) : garde-fous anti-régression temporelle (`OrderTimestamp`) et anti-double-application. Couleur choisie aléatoirement parmi 17 couleurs Material à la création.

#### Machine à états pilotée par événements

`PostSend` → `message.api.sent` → stocké `pending` → dispatché (immédiat ou différé) → `GetOutstanding` → `message.phone.sending` → `sending` → `PostEvent` → `message.phone.sent`/`delivered`/`message.send.failed` → statuts terminaux. En parallèle : notification FCM → `message.notification.sent` → `message.send.expired.check` programmé → si toujours pending à l'échéance, `message.send.expired` → retry si `SendAttemptCount < MaxSendAttempts`.

`MessageListener` route 12 types d'événements. `onMessageNotificationFailed` (échec FCM) route vers le **même** chemin que l'échec SMS réel rapporté par le téléphone.

### Incohérences détectées

1. **[services/message_service.go:710-714 vs 648-651, 786-789] Gestion incohérente des statuts inattendus** — `HandleMessageDelivered` avale silencieusement (`return nil`), `HandleMessageSent`/`HandleMessageExpired` retournent une erreur pour un cas similaire.

2. **[services/message_service.go:643-651 vs 683-686] Idempotence incohérente** — `HandleMessageSent` protège contre `IsSent()||IsDelivered()`, `HandleMessageFailed` ne protège que contre `IsDelivered()` : un message déjà `sent` peut repasser `failed` sur un event tardif/dupliqué.

3. **[services/message_service.go:1037-1042] `enrichErrorMessage` fragile** — détection par `strings.Contains` d'une seule erreur Android connue, non extensible.

4. **[services/message_thread_service.go:212-234] `getColor()` recrée un `rand.New` à chaque appel** — collision de seed possible en concurrence (bulk-send), même couleur pour threads différents.

5. **[services/message_thread_service.go:138-142 vs message_service.go:882-885] Thread manquant non géré comme cas bénin**, contrairement à `CheckExpired` qui traite `ErrCodeNotFound` explicitement comme "déjà supprimé, ignorer".

6. **[services/message_thread_service.go:144-149] Deux styles de gestion d'erreur dans la même fonction** pour des échecs DB de même nature.

7. **[SÉCURITÉ] v2 `POST /v2/send` sans vérification de possession du numéro `from`** — contrairement à tous les endpoints v1 protégés par `authorizePhoneAPIKey`, v2 prend `Owner` directement du body sans vérifier via `PhoneRepository` qu'il appartient à l'utilisateur authentifié.

8. **[handlers/v2_message_handler.go:106-111] Code d'erreur `EMPTY_MESSAGE` réutilisé à tort** pour "message trop long" (>1600 caractères) — devrait être un code distinct.

9. **[v2_message_handler.go:127-130 vs message_handler.go:107-112] Sémantique HTTP divergente** pour le même échec de service : v1 → 500 standard, v2 → 502 + format `{"error"}`.

10. **Documentation Swagger incomplète** — `PostReceive` ne documente pas `@Failure 401` alors qu'il appelle `authorizePhoneAPIKey`, contrairement à des endpoints équivalents.

11. **Coquille de fichier** — `events/message_thead_api_deleted_event.go` (`thead` au lieu de `thread`).

12. **Typos répétés "thread thread"** dans les logs/erreurs de `message_thread_handler.go` (lignes 168, 172, 178, 183).

13. **Duplication complète de la logique bulk-send** entre `MessageHandler.BulkSend` (JSON) et `BulkMessageHandler.Store` (CSV) — validation, billing, goroutines, agrégation, tout dupliqué.

14. **`entities/message.go` sans `IsReceived()`/`IsDeleted()`** alors que tous les autres statuts ont leur helper.

15. **Conflation échec SMS / échec notification push** — `onMessageNotificationFailed` route vers le même `MessageEventNameFailed` que l'échec réel d'envoi rapporté par le téléphone ; pas de distinction de cause dans `FailureReason`.

---

## 4. Domaine : Billing, Apps & LemonSqueezy

### Documentation

#### Entité `App`

Intégration tierce (pas un téléphone Android) : `APIKey` (`ak_`), `APISecret` (`sk_`, `json:"-"`), `PhoneNumbers` autorisés comme expéditeur, `WebhookURL`/`WebhookSecret` optionnels.

#### `BillingUsage`

Un enregistrement par utilisateur/mois calendaire. `IsEntitled(count, limit)` : `TotalMessages()+count < limit` (strictement inférieur).

#### Endpoints `AppHandler`

- `RegisterRoutes` (`AuthenticatedMiddleware` seul) : CRUD `/v1/apps`.
- `RegisterAppKeyRoutes` (`AppAuthMiddleware` + `AuthenticatedMiddleware`) : `/v1/apps/messages/send`, `/status` — réservées à l'auth `ak_` (actuellement cassée, voir Auth #1).
- `sendMessage` : refuse sans `phone_numbers` (422), vérifie billing (402 si dépassement), prend systématiquement `PhoneNumbers[0]`.

#### Middleware `AppAuth`

`x-api-key` (`ak_`) + `x-api-secret`, comparaison `crypto/subtle.ConstantTimeCompare` (bon réflexe anti-timing-attack).

#### `BillingHandler`

`GET /v1/billing/usage` (mois courant), `/usage-history` (mois passés). `AuthenticatedMiddleware` seul — accepte session, JWT, `uk_`, `ak_`, ou `pk_`.

#### `BillingService`

Alerte email à seuils fixes selon le plan (ex: Free à 160/180/190 sur 200), déduplication Redis 12h. `RegisterSentMessage`/`RegisterReceivedMessage` incrémentent puis vérifient les seuils à chaque message.

#### Plans (`entities/user.go:17-114`)

`free`, `pro-monthly/yearly/lifetime`, `ultra-monthly/yearly`, `20k-monthly/yearly`, `100k-monthly`, `50k-monthly`, `200k-monthly` (**pas de yearly pour ces 3 derniers**). Limites : free=200, pro=5000, ultra=10000, 20k=20000, 50k=50000, 100k=100000, 200k=200000, défaut=200.

#### Intégration LemonSqueezy

`POST /lemonsqueezy/event` — **sans aucun middleware**, seule protection = signature HMAC (`X-Signature`). 4 events gérés (`subscription_created/cancelled/expired/updated`). Résolution utilisateur via `CustomData["user_id"]` ou fallback email/subscription_id. `subscriptionName(variant)` mappe par recherche de sous-chaînes.

### Incohérences détectées

1. **[CRITIQUE] Plans annuels 50k/100k/200k inexistants.** `entities/user.go:58-71` ne définit aucune constante yearly pour ces plans (seul `20KYearly` existe). `lemonsqueezy_service.go:215-238` : les branches ne testent que "monthly" → un plan annuel de ces familles tombe sur `return SubscriptionNameFree` (ligne 240). **Un client payant est classé "free".**

2. **`LEMONSQUEEZY_API_KEY`/`LEMONSQUEEZY_SIGNING_SECRET` absents de `.env.docker`** — même classe de bug que `EVENTS_QUEUE_USER_ID` déjà corrigé. En dev, `WithSigningSecret("")` — impact exact à vérifier selon l'implémentation de la lib.

3. **`sendMessage` ignore le numéro expéditeur demandé** — prend inconditionnellement `PhoneNumbers[0]`, aucun moyen pour le client de choisir si l'App a plusieurs numéros.

4. **`AppStoreRequest.WebhookSecret` jamais validé** — possible de créer une App avec `WebhookURL` sans secret.

5. **`IsEntitledWithCount` "fail open"** — une panne DB (Load ou GetCurrent) retourne `nil` (autorisé), levant silencieusement toutes les limites d'envoi en cas d'incident infra.

6. **`gormBillingUsageRepository` utilise `crdbgorm` (CockroachDB)** sur un dev Postgres — à vérifier si la prod tourne réellement sur CockroachDB.

7. *(Non-problème confirmé)* `responsePaymentRequired` (402) est correctement utilisé pour les dépassements de quota — bonne pratique HTTP, mentionné pour contraste avec d'autres endroits de l'API moins cohérents.

---

## 5. Domaine : Webhooks, Events & Intégrations

### Documentation

#### Le pipeline d'événements (CloudEvents)

Toute la logique asynchrone (facturation, notifications, webhooks, websocket, heartbeats) repose sur un bus d'événements interne au format CloudEvents.

- `EventDispatcher.Dispatch`/`DispatchWithTimeout` : valide l'event, construit une tâche avec header `x-api-key: <EVENTS_QUEUE_USER_API_KEY>`, l'empile via `queue.Enqueue`.
- Fallback local en mémoire si `Enqueue` retourne `DeadlineExceeded` (ne se déclenche qu'en prod avec Cloud Tasks ; mort en dev avec l'émulateur).
- `Publish` exécute en parallèle tous les listeners abonnés ; si aucun listener n'est abonné, log INFO et **abandon silencieux**.
- **Boucle interne HTTP `/v1/events`** : en dev, l'émulateur POST le body vers son propre `EVENTS_QUEUE_ENDPOINT` (le serveur s'auto-appelle). En prod, Google Cloud Tasks avec 3 tentatives.
- `EventsHandler.Dispatch` : parse CloudEvents, **vérifie `AuthContext.ID == queueConfig.UserID` exactement**, sinon 403. Auth via `AuthenticatedMiddleware` seul (session/JWT/`uk_` — **pas** `pk_`, confirmé en session).
- Config lue directement des env vars **sans validation au démarrage**.

#### Catalogue des événements (31 au total)

Convention de nommage incohérente : majorité `EventType*`, 11 sans préfixe. Table complète des associations event→listener dans le fichier source du fork (`domain-webhooks-events.md`), résumé : la quasi-totalité des events métier ont au moins un listener actif, à l'exception notable de ceux ciblant `phone_notification_listener` (mort, voir incohérence #1) et de `user.api-key.rotated` (listener présent mais jamais dispatché, incohérence #6).

#### Listeners enregistrés (13 sur 14 fichiers)

`container.go` enregistre 13 listeners. **`phone_notification_listener.go` (`NewNotificationListener`) n'est appelé nulle part.**

#### Webhooks — deux systèmes parallèles

**V1/legacy** : CRUD complet `/v1/webhooks`, limite 10/utilisateur, signature **JWT HS256** (ne hash pas le payload, prouve juste la connaissance de la clé), retry 2 tentatives/5s. Cas spécial Discord intégré dans la logique générique de payload.

**V2** : pas de CRUD (champs directs sur App/User), priorité App puis fallback User, signature **HMAC-SHA256** du payload (plus robuste), retry 3 tentatives backoff fixe [1s,5s,30s] en goroutine détachée sur `context.Background()` (perte silencieuse si le process redémarre pendant une tentative), ne couvre que 3 events contre ~10+ pour le V1.

#### Discord (intégration native)

CRUD `/v1/discord-integrations` (max 1/utilisateur) + `/discord/event` public (signature Ed25519, handshake PING, Slash Commands). `sendSMS` traite tout en synchrone dans le handler HTTP (validation + billing + envoi), contrairement au pattern validate→dispatch→listener du reste de l'API.

#### Intégration 3CX

`POST /integration/3cx/messages`, 4 des 5 handlers du listener commentés (code mort non supprimé), `spew.Dump` inconditionnel du corps de chaque requête en logs.

### Incohérences détectées

1. **[CRITIQUE] `phone_notification_listener.go` jamais enregistré** — si rien ne duplique ce rôle, les notifications FCM déclenchant l'envoi SMS réel (`message.api.sent`, `message.send.retry`, `message.notification.send`) et les alertes heartbeat manqué (`phone.heartbeat.missed`) ne partent jamais. **Vérification urgente recommandée.**

2. **[CRITIQUE, déjà vécu] `/v1/events` exige un match exact d'UserID sans validation au démarrage** — une mauvaise config (placeholder, mauvais type de clé) fait échouer tout le pipeline silencieusement, sans health-check ni alerte.

3. **`emulatorPushQueue.Enqueue` sans retry ni remontée d'erreur** — contrairement à la queue Cloud Tasks (3 tentatives), un event perdu en dev est indétectable sans grep manuel.

4. **Deux systèmes de webhooks incompatibles** (V1 JWT/10 events/retry 2x5s vs V2 HMAC/3 events/retry backoff) — comportement de sécurité incohérent selon le canal.

5. **Message d'erreur incohérent avec la limite réelle** — limite webhooks codée à 10, message utilisateur dit "10", mais le log interne (`ctxLogger.Warn`) dit "5" (commentaire non mis à jour).

6. **`user.api-key.rotated` a un listener mais rien ne semble le dispatcher** — à vérifier dans `user_handler.go`/`user_service.go` (non lus intégralement dans cet audit).

7. **4/5 handlers du listener 3CX désactivés par commentaire**, code mort maintenu non supprimé.

8. **`spew.Dump` inconditionnel du corps des requêtes 3CX** en logs, y compris en production — fuite potentielle de contenu SMS.

9. **Typo utilisateur-visible** — `"TIMOUT after 10 seconds"` dans `webhook_service.go:370`, potentiellement visible dans des events `webhook.send.failed`.

10. **Convention de nommage des constantes d'event incohérente** (`EventType*` vs sans préfixe, 11 exceptions sans règle apparente).

11. **Discord traite l'envoi SMS en synchrone dans le handler HTTP**, seul canal à déroger au pattern validate→dispatch→listener asynchrone du reste de l'API.

---

## Méthodologie

Documentation produite par 5 agents en parallèle, chacun ayant lu intégralement les fichiers de son périmètre (handlers, services, repositories, entities, middlewares, listeners, validators, requests) plutôt que des extraits. Les fichiers de requests/validators de routine (DTO simples) ont été identifiés par nom mais pas systématiquement lus ligne à ligne dans le domaine Messages (volume trop important) — signalé explicitement dans ce cas.

Aucune incohérence listée ici n'a été corrigée automatiquement — ce document est un audit, pas un correctif. Voir le résumé exécutif en tête de document pour l'ordre de priorité recommandé.
