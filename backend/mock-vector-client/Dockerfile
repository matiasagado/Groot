# Use the official Vector image
FROM timberio/vector:latest-alpine

# Create a directory for the configuration
WORKDIR /etc/vector

# Copy the Vector configuration template
COPY vector-config.yaml.template /etc/vector/vector-config.yaml.template

# Install gettext for envsubst
RUN apk update --no-cache && \
    apk add --no-cache gettext && \
    # Remove the default config file
    rm -f /etc/vector/vector.yaml

# Create the entrypoint script
RUN printf '#!/bin/sh\nenvsubst < /etc/vector/vector-config.yaml.template > /etc/vector/vector.yaml\ncat /etc/vector/vector.yaml\nexec vector --config /etc/vector/vector.yaml\n' > /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Set the default HTTP sink URI (can be overridden)
ENV HTTP_SINK_URI="http://poc-log-preprocessor:9501/test/http_sink"

# Use the custom entrypoint
ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
