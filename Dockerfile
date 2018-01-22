FROM scratch
ADD prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
VOLUME [ "/etc/prometheus-targets" ]
CMD [""]
