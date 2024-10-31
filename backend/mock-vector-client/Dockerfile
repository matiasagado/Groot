# Use the official Vector image
FROM timberio/vector:latest-alpine

# Create a directory for the configuration
WORKDIR /etc/vector

# Copy the Vector configuration template
COPY vector-config.yaml.template /etc/vector/vector-config.yaml.template

# Install gettext-base for envsubst
RUN apk add update --no-cache && \
    apk add gettext

# Create the entrypoint script
RUN echo '#!/bin/bash\n\
envsubst < /etc/vector/vector-config.yaml.template > /etc/vector/vector.yaml\n\
exec vector --config /etc/vector/vector.yaml\n'\
> /entrypoint.sh && \
chmod +x /entrypoint.sh

# Set the default HTTP sink URI (can be overridden)
ENV HTTP_SINK_URI="http://poc-log-preprocessor:9501/test/http_sink"

# Use the custom entrypoint
ENTRYPOINT ["/entrypoint.sh"]
