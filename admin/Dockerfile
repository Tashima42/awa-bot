FROM node:lts-alpine as build-stage
RUN mkdir -p /app
WORKDIR /app
COPY ./admin/package*.json ./
RUN npx --yes yarn install
COPY ./admin .
RUN npx --yes yarn run build

FROM nginx:stable-alpine as production-stage
EXPOSE 8080
COPY ./admin/nginx.conf /etc/nginx/conf.d/default.conf
RUN rm -rf /usr/share/nginx/html/*
COPY --from=build-stage /app/dist /usr/share/nginx/html
CMD ["nginx", "-g", "daemon off;"]
