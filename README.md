# riemann-windows-service-check

Simple windows command line binary to query service state of remote hosts and send the value as an event to riemann




## config

Create a file named `config.yaml` with this format:


    riemann_host: <your riemann host>:5555
    event_ttl: 300  # TTL before riemann expires the event>
    check_timeout_seconds: 30  # Seconds before the service check gives up and reports the host down/unreachable
    services:
      <service name>:
        - host1
        - host2
      <service name 2>:
        - host1
        - host2
        - host3
