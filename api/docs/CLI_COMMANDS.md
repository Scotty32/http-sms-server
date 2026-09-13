# CLI Commands

Toutes les commandes CLI du projet vivent sous `api/cmd/`. Elles utilisent le même
`di.NewContainer(...)` que le serveur HTTP, donc elles ont besoin des mêmes variables
d'environnement (`.env` en local, `.env.docker` dans le conteneur dev).

En développement, elles s'exécutent directement dans le conteneur `httpsms-api`
(qui a le toolchain Go complet via `Dockerfile.dev`), sans rien construire :

```bash
docker exec httpsms-api go run ./cmd/<nom-de-la-commande> [arguments]
```

> Ces commandes ne sont **pas** buildées dans l'image de production (`Dockerfile` ne
> compile que `main.go` à la racine) — elles sont réservées à l'usage admin/ops via un
> accès shell au conteneur (ou en local avec `go run` depuis `api/`).

---

## `cmd/plans` — Gestion des plans tarifaires et attribution aux utilisateurs

Gère le catalogue `entities.Plan` (table `plans`) et permet d'attribuer un plan à un
utilisateur. C'est la seule commande de ce back-office minimal : il n'existe pas
d'interface web pour ça, uniquement le CLI (accès admin = accès shell au conteneur).

### `seed` — Créer/mettre à jour les plans depuis un fichier JSON (idempotent)

```bash
docker exec httpsms-api go run ./cmd/plans seed [--file=cmd/plans/plans.json]
```

- `--file` (optionnel, défaut `cmd/plans/plans.json`) : chemin du catalogue JSON à charger.
- Idempotent : fait un `UPSERT` sur la colonne `name` (clé unique). Relancer la commande
  avec un fichier modifié met à jour les plans existants sans dupliquer de lignes.
- Format d'une entrée du fichier JSON :
  ```json
  {
    "name": "pro-monthly",
    "display_name": "Pro (Monthly)",
    "message_limit": 5000,
    "phone_limit": 3,
    "price_cents": 999,
    "currency": "USD",
    "billing_period": "monthly"
  }
  ```
- Le plan `free` doit toujours exister avec `message_limit: 5` et `phone_limit: 1` — ce
  sont les limites appliquées à tout utilisateur sans abonnement payant (voir
  [BUSINESS_LOGIC_AUDIT.md](./BUSINESS_LOGIC_AUDIT.md)).

### `list` — Lister tous les plans actuellement en base

```bash
docker exec httpsms-api go run ./cmd/plans list
```

Affiche une ligne par plan : `name`, `display_name`, `message_limit`, `phone_limit`,
`price_cents`/`currency`/`billing_period`.

### `assign` — Attribuer un plan à un utilisateur

```bash
docker exec httpsms-api go run ./cmd/plans assign --email=user@example.com --plan=pro-monthly
```

- `--email` (obligatoire) : email de l'utilisateur cible.
- `--plan` (obligatoire) : `name` d'un plan existant (voir `list`). La commande échoue si
  le plan n'existe pas — lancer `seed` d'abord.
- Met à jour `users.subscription_name`. Les nouvelles limites (messages/mois, numéros de
  téléphone) s'appliquent immédiatement, sans redémarrage de l'API.

---

## `cmd/loadtest` — Test de charge / envoi manuel de SMS

```bash
docker exec httpsms-api go run ./cmd/loadtest
```

- Pas de flags : le comportement est câblé en dur dans `main()` (actuellement appelle
  `sendSingle()`, `bulkSend()` est disponible mais non appelée par défaut — à commenter/
  décommenter dans le code selon le besoin).
- Variables d'environnement utilisées :
  - `HTTPSMS_KEY` / `HTTPSMS_KEY_BULK` — clé API (`x-api-key`) à utiliser.
  - `HTTPSMS_FROM` / `HTTPSMS_FROM_BULK` — numéro expéditeur.
  - `HTTPSMS_TO` / `HTTPSMS_TO_BULK` — numéro(s) destinataire(s).
  - `HTTPSMS_ENCRYPTION_KEY` — utilisée par `decode()` (chiffrement E2E des messages).
- Envoie une requête réelle contre `http://sms-dev.om-ci.org` (`/v1/messages/send` ou
  `/v1/messages/bulk-send`) — **pas un mock**, ça envoie un vrai SMS via le pipeline
  complet si les identifiants sont valides.

## `cmd/fcm` — Test d'envoi d'une notification push Firebase (heartbeat)

```bash
docker exec httpsms-api go run ./cmd/fcm
```

- Pas de flags.
- Variable d'environnement requise : `FIREBASE_TOKEN` (token FCM d'un appareil Android
  cible).
- Envoie un message data-only `{"KEY_HEARTBEAT_ID": "<timestamp RFC3339>"}` avec
  `priority: high` — utile pour vérifier que le projet Firebase configuré côté API
  correspond bien à celui de l'app Android (cause d'un bug corrigé cette session : les
  IDs de projet Firebase de l'API et de l'app ne correspondaient pas, donc aucun push
  n'arrivait jamais).

## `cmd/replay` et `cmd/migration` — Stubs, non implémentés

```bash
docker exec httpsms-api go run ./cmd/replay
docker exec httpsms-api go run ./cmd/migration
```

Ces deux commandes se contentent de charger `.env` et (pour `replay`) d'instancier le
container DI, sans aucune action derrière. Ce sont des emplacements réservés pour de la
logique future (rejouer des événements CloudEvents / migrations de données manuelles) —
il n'y a rien à en attendre en l'état actuel du code.

---

## Le serveur lui-même (`api/main.go`)

Ce n'est pas une commande d'admin mais le point d'entrée du serveur HTTP :

```bash
go run . # ou le binaire compilé, avec les variables d'environnement du .env chargées
```

- Si lancé sans argument (`len(os.Args) == 1`), charge automatiquement `.env` via
  `di.LoadEnv()`. En conteneur Docker, les variables viennent de `env_file:` dans
  `docker-compose.dev.yml`, pas de `.env`.
- Si `EVENTS_QUEUE_TYPE=redis`, démarre en plus un worker Redis (`asynq`) en arrière-plan
  pour consommer la queue d'événements internes.
- Écoute sur `${APP_HOST}:${APP_PORT}`.

**Rappel important (déjà rencontré plusieurs fois cette session) :** `docker restart
httpsms-api` ne recharge PAS les variables d'un `env_file:` modifié. Après un changement
dans `.env.docker`, il faut forcer la recréation du conteneur :

```bash
docker compose -f docker-compose.dev.yml up -d --force-recreate api
```
