# build stage
FROM node:lts-alpine as build

WORKDIR /app

COPY package.json package-lock.json ./

# install python
RUN apk add --no-cache python3

RUN npm install
COPY . .
RUN npm run generate

# production stage
FROM nginx:stable-alpine as production
COPY --from=build /app/dist /usr/share/nginx/html

# Copy the nginx configuration file
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
