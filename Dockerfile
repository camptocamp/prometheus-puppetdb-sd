FROM scratch
ADD puppetdb-prometheus /
ADD prometheus.yml /etc/prometheus-config/
ENTRYPOINT ["/puppetdb-prometheus"]
CMD [""]
