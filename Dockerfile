FROM registry.access.redhat.com/ubi7/ubi-minimal:latest
COPY cluster-service . 
ENTRYPOINT [ "./cluster-service" ]
