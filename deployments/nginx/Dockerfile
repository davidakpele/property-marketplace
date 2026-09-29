FROM nginx:1.27-alpine

COPY nginx.conf      /etc/nginx/nginx.conf
COPY conf.d/         /etc/nginx/conf.d/

RUN mkdir -p /var/cache/nginx/api /var/cache/nginx/tmp \
    && chown -R nginx:nginx /var/cache/nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
