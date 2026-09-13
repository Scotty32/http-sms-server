# build stage
FROM node:lts-alpine as build

WORKDIR /app

# install python (needed by some native module builds) and pnpm
RUN apk add --no-cache python3 && npm install -g pnpm

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run generate

# production stage
FROM nginx:stable-alpine as production
COPY --from=build /app/.output/public /usr/share/nginx/html

# Copy the nginx configuration file
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
