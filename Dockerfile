FROM ubuntu:22.04

ARG BIN_PATH="./bin"
ARG BIN_NAME="aws-disaster-recovery-tool"
ARG APP_DIR="/opt"
ARG ENV_PRD="./.env.production"
ARG ENV_BKP="./.env.backup"



COPY "${BIN_PATH}/${BIN_NAME}" $APP_DIR
COPY $ENV_BKP $APP_DIR
COPY $ENV_PRD $APP_DIR

RUN apt update && \
    apt install ca-certificates -y

WORKDIR $APP_DIR
  
ENTRYPOINT ["/opt/aws-disaster-recovery-tool", "-e", "production"]
