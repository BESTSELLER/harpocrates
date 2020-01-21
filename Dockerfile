FROM alpine
WORKDIR /
COPY ./harpocrates /

COPY docker-entrypoint.sh /
RUN chmod +x "/docker-entrypoint.sh"
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]