# Déploiement en production sur Dokploy

Ce guide décrit comment déployer httpSMS (API Go + Web Nuxt + Postgres + Redis) sur
Dokploy avec `docker-compose.dokploy.yml`, qui réutilise tel quel les Dockerfiles de
production existants (`api/Dockerfile`, `web/Dockerfile`).

## ⚠️ À faire avant tout déploiement

`web/.env.production` et `api/.env.docker` ont été nettoyés (les vraies clés du SaaS
`httpsms.com` et la clé privée de service account Firebase exposée ont été remplacées
par des placeholders vides / `xxxxx`). Il reste à :

1. Remplir `web/.env.production` avec la config web SDK de **votre propre** projet
   Firebase et vos domaines (voir étape 3 ci-dessous).
2. Ne jamais remplir `api/.env.docker` avec de vraies valeurs — ce fichier est pour le
   dev local uniquement ; en production, les secrets (dont `FIREBASE_CREDENTIALS`)
   passent exclusivement par l'onglet Environment de Dokploy (étape 2).
3. **Révoquer/régénérer** l'ancienne clé de service account Firebase
   (`firebase-adminsdk-fbsvc@httpsms-ba994...`) dans la console GCP si ce projet est
   encore actif : elle reste présente dans l'historique Git même après ce nettoyage,
   tant qu'elle n'a pas été réécrite (`git filter-repo`/BFG) ou que le repo n'est pas
   privé et déjà poussé. Éditer le fichier ne suffit pas à invalider la clé.

## Prérequis

- Une instance Dokploy avec accès à ce repo Git (public ou via une clé de déploiement).
- Un projet Firebase (pour FCM + auth) — voir la section "Setup Firebase" du `README.md`.
- Un service SMTP (Mailtrap, SendGrid, etc.) pour l'envoi d'emails.
- Deux sous-domaines pointant vers votre serveur Dokploy, par ex. :
  - `app.votredomaine.com` → interface web
  - `api.votredomaine.com` → API

## 1. Créer l'application dans Dokploy

1. Dans Dokploy : **Create Project** → **Add Service** → **Compose**.
2. Connectez le repo Git et la branche à déployer (`main`).
3. Dans **Compose Path**, indiquez `docker-compose.dokploy.yml` (au lieu du
   `docker-compose.yml` par défaut, qui est prévu pour du dev/test local avec ports
   exposés).

## 2. Configurer les variables d'environnement (onglet Environment)

Dokploy écrit ces valeurs dans un `.env` à côté du compose file, utilisé pour
l'interpolation `${VAR}`. Renseignez :

```dotenv
# Postgres
POSTGRES_DB=httpsms
POSTGRES_USER=dbusername
POSTGRES_PASSWORD=<mot de passe fort, généré aléatoirement>

# Firebase (Admin SDK — voir README "1. Setup Firebase")
GCP_PROJECT_ID=<votre-projet-firebase>
FIREBASE_CREDENTIALS=<contenu JSON complet du firebase-credentials.json, sur une seule ligne>

# Utilisateur système (voir étape 5 ci-dessous)
EVENTS_QUEUE_USER_API_KEY=<généré à l'étape 5>
EVENTS_QUEUE_USER_ID=<généré à l'étape 5>

# SMTP
SMTP_FROM_EMAIL=noreply@votredomaine.com
SMTP_USERNAME=<...>
SMTP_PASSWORD=<...>
SMTP_HOST=<...>
SMTP_PORT=587

# URLs publiques
APP_URL=https://app.votredomaine.com
SWAGGER_HOST=api.votredomaine.com

# Secrets applicatifs — générez des chaînes aléatoires longues, ex: `openssl rand -hex 32`
SMS_WEBHOOK_SECRET=<random>
JWT_SECRET=<random>
```

Variables optionnelles (laissez vides si non utilisées) : `UPTRACE_DSN`,
`PUSHER_APP_ID`, `PUSHER_KEY`, `PUSHER_SECRET`, `PUSHER_CLUSTER`.

> Ne mettez **aucun** de ces secrets dans un fichier committé — c'est le rôle de
> l'onglet Environment de Dokploy.

## 3. Configurer `web/.env.production` (config publique du frontend)

Ce fichier est buildé dans le bundle statique par `nuxi generate` (Nuxt charge
`.env.production` automatiquement en mode production) — ce ne sont **pas** des
secrets serveur, mais la config publique du client (clé web Firebase, URL de l'API,
etc.), donc il est normal qu'il soit versionné. Éditez-le et commitez vos propres
valeurs avant de déployer :

```dotenv
API_BASE_URL=https://api.votredomaine.com
APP_URL=https://app.votredomaine.com
APP_NAME=httpSMS
APP_ENV=production

FIREBASE_API_KEY=...
FIREBASE_AUTH_DOMAIN=...
FIREBASE_PROJECT_ID=...
FIREBASE_STORAGE_BUCKET=...
FIREBASE_MESSAGING_SENDER_ID=...
FIREBASE_APP_ID=...
FIREBASE_MEASUREMENT_ID=...
```

(Supprimez ou laissez vides `CHECKOUT_URL` / `ENTERPRISE_CHECKOUT_URL` /
`CLOUDFLARE_TURNSTILE_SITE_KEY` si vous n'utilisez pas la facturation Lemonsqueezy du
SaaS d'origine.)

## 4. Premier déploiement

Lancez le déploiement depuis Dokploy (**Deploy**). Il va builder `api/Dockerfile` et
`web/Dockerfile`, créer les volumes `httpsms_postgres_data`, `httpsms_redis_data`,
`httpsms_api_logs`, et démarrer les 4 services sur le réseau interne + `dokploy-network`.

## 5. Créer l'utilisateur système

L'API a besoin d'un "system user" en base pour traiter les événements async
(`EVENTS_QUEUE_USER_ID` / `EVENTS_QUEUE_USER_API_KEY`). Une fois le stack démarré :

1. Générez un UUID et une clé API aléatoire pour ce user.
2. Ouvrez un shell Postgres depuis Dokploy (onglet **Terminal** du service `postgres`,
   ou `docker exec -it <container_postgres> psql -U dbusername -d httpsms`) :

   ```sql
   INSERT INTO users (id, api_key, email) VALUES ('<uuid-généré>', '<clé-générée>', 'system@votredomaine.com');
   ```

3. Reportez ces mêmes valeurs dans `EVENTS_QUEUE_USER_ID` et
   `EVENTS_QUEUE_USER_API_KEY` (onglet Environment, étape 2).
4. Redéployez/redémarrez le service `api` pour qu'il prenne en compte ces variables.

## 6. Domaines et HTTPS

Dans Dokploy, onglet **Domains** du service Compose :

- Ajoutez `api.votredomaine.com` → service `api`, port conteneur `8000`, HTTPS activé
  (Let's Encrypt automatique).
- Ajoutez `app.votredomaine.com` → service `web`, port conteneur `3000`, HTTPS activé.

Dokploy génère lui-même les labels Traefik nécessaires — aucune modification du
compose file n'est requise pour ça.

## 7. Vérification

- `https://api.votredomaine.com/health` (ou l'endpoint de santé exposé par l'API)
  répond correctement.
- `https://app.votredomaine.com` charge l'interface et permet de créer un compte.
- Les logs de l'API sont visibles à la fois via `docker logs` (Dokploy → onglet Logs)
  et persistés dans le volume `httpsms_api_logs` (`/app/logs/api.log`).

## 8. Build de l'app Android

Une fois l'API en prod, régénérez `android/app/google-services.json` avec votre
projet Firebase, mettez à jour l'URL de l'API dans l'app, et buildez l'APK (voir
README section 7).

## Notes de sécurité

- `postgres` et `redis` ne publient aucun port sur l'hôte dans ce compose — ils ne
  sont joignables que via le réseau interne `internal`, pas via `dokploy-network`.
- `WEBHOOK_TEST_ALLOW_PRIVATE_IPS` est forcé à `false` (contrairement au
  `docker-compose.dev.yml` qui l'active pour tester les webhooks entre conteneurs) —
  ne pas repasser à `true` en prod, ça ouvrirait une SSRF sur le endpoint de test webhook.
- Pas de `container_name` explicite dans le compose, pour laisser Dokploy gérer le
  nommage sans collision si vous déployez plusieurs environnements (staging/prod).
