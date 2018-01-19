FROM scratch
ADD prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
CMD [""]
