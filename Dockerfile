FROM scratch
ADD prometheus-puppetdb /
ADD prometheus.yml /etc/prometheus-config/
ENTRYPOINT ["/prometheus-puppetdb"]
VOLUME [ "/etc/prometheus-config" ]
CMD [""]
