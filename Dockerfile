FROM alpine
WORKDIR /
COPY /tmp/harpocrates /

COPY docker-entrypoint.sh /
RUN chmod +x "/docker-entrypoint.sh"
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]