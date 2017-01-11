FROM scratch
ADD puppetdb-prometheus /
ENTRYPOINT ["/puppetdb-prometheus"]
CMD [""]
