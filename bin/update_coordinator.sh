#!/usr/bin/env bash

if [[ "${REEVE_TAG_PRESTO_ROLE}" != "coordinator" ]]; then
  echo "$0 is only to be executed at the Reeve Coordinator." >&2
  exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

source ${SCRIPT_DIR}/../config/global.conf
[[ -e ${SCRIPT_DIR}/../config/coordinator.conf ]] && source ${SCRIPT_DIR}/../config/coordinator.conf

restart=false

mkdir -p ${DOWNLOADS_DIR}

prestoversion="${PRESTO_SERVER_RPM_VERSION}"
prestourl="${PRESTO_SERVER_URL}"
prestomd5sum="${PRESTO_SERVER_MD5SUM}"
javaversion="${JAVA_8_APP_VERSION}"
javaurl="${JAVA_8_URL}"
javamd5sum="${JAVA_8_MD5SUM}"

javainstalledversion=$(java -version 2>&1 | head -1 | sed 's/.*"\(.*\)".*/\1/')
if [[ ${javainstalledversion} != ${javaversion} ]]; then
  rpmaction="-i"
  [[ -n ${javainstalledversion} ]] && rpmaction="-U"

  echo "Downloading Java Oracle v${javaversion}"
  filename=${javaurl##*/}
  curl -s -L --header "Cookie: gpw_e24=http%3A%2F%2Fwww.oracle.com%2F; oraclelicense=accept-securebackup-cookie" \
    ${javaurl} -o ${DOWNLOADS_DIR}/${filename}
  if ! echo "${javamd5sum}  ${DOWNLOADS_DIR}/${filename}" | md5sum -c --quiet - &>/dev/null; then
    echo "Checksum for ${filename} do NOT match" >&2
    exit 1
  fi
  echo "Updating/Installing Java Oracle v${javaversion}"
  rpm ${rpmaction} ${DOWNLOADS_DIR}/${filename} 2>&1 >/dev/null
  restart=true
else
  echo "Java Oracke v${javainstalledversion} exists and is the latest version"
fi

msgaction="install"
if [[ -x /root/bin/presto ]]; then
  msgaction="update"
  prestocliinstalledversion=$(/root/bin/presto --version)
  if [[ ${prestocliinstalledversion} != "Presto CLI ${PRESTO_CLI_VERSION}" ]]; then
    echo "Downloading Presto CLI v${PRESTO_CLI_VERSION}"
    filename=${PRESTO_CLI_URL##*/}
    curl -s ${PRESTO_CLI_URL} -o ${DOWNLOADS_DIR}/${filename}
    if ! echo "${PRESTO_CLI_MD5SUM}  ${DOWNLOADS_DIR}/${filename}" | md5sum -c --quiet - &>/dev/null; then
      echo "Checksum for ${filename} do NOT match" >&2
      exit 1
    fi
    echo "Updating/Installing Presto CLI v${PRESTO_CLI_VERSION}"
    mkdir -p /root/bin
    \cp ${DOWNLOADS_DIR}/${filename} /root/bin/presto
    chmod +x /root/bin/presto

    # Verify Presto CLI was installed sucessfully
    prestocliinstalledversion=$(/root/bin/presto --version)
    if [[ ${prestocliinstalledversion} != "Presto CLI ${PRESTO_CLI_VERSION}" ]]; then
      echo "Presto CLI ${msgaction} fail. (${prestocliinstalledversion} != ${PRESTO_CLI_VERSION})" >&2
      exit 1
    fi
  else
    echo "Presto CLI v${prestocliinstalledversion} exists and is the latest version"
  fi
fi



prestoinstalledversion=$(rpm -aq | grep presto | sed 's/presto-server-rpm-\(.*\).x86_64/\1/')
if [[ ${prestoinstalledversion} != ${prestoversion} ]]; then
  rpmaction="-i"
  [[ -n ${prestoinstalledversion} ]] && rpmaction="-U"

  echo "Downloading Presto Server v${prestoversion}"
  filename=${prestourl##*/}
  curl -s ${prestourl} -o ${DOWNLOADS_DIR}/${filename}
  if ! echo "${prestomd5sum}  ${DOWNLOADS_DIR}/${filename}" | md5sum -c --quiet - &>/dev/null; then
    echo "Checksum for ${filename} do NOT match" >&2
    exit 1
  fi
  echo "Updating/Installing Presto Server v${prestoversion}"
  rpm ${rpmaction} ${DOWNLOADS_DIR}/${filename} 2>&1 >/dev/null
  restart=true

  # # prestoversion=$(echo $prestourl | sed 's|.*/presto/\(.*\)/presto-server.*|\1|')
  # prestoinstalledversion=$(rpm -aq | grep presto | sed 's/presto-server-rpm-\(.*\).x86_64/\1/')
  # echo "Presto Server v${prestoinstalledversion} sucessfully updated at the Coordinator"

  echo "Updating Presto Server at the workers"
  serf event presto/update
else
  echo "Presto Server v${prestoinstalledversion} exists and is the latest version"
fi

httpport=${HTTP_SERVER_PORT}
querymaxmem=${QUERY_MAX_MEMORY}
querymaxmempernode=${QUERY_MAX_MEMORY_PER_NODE}
coordinatorIP=`serf members | grep 'presto_role=coordinator' | awk '{print \$2}' | cut -f1 -d:`
discoveryuri="http://${coordinatorIP}:${httpport}"

cat <<-EOCONFW > /etc/presto/config.properties.new
coordinator=true
node-scheduler.include-coordinator=false
http-server.http.port=${httpport}
query.max-memory=${querymaxmem}
query.max-memory-per-node=${querymaxmempernode}
discovery-server.enabled=true
discovery.uri=${discoveryuri}
EOCONFW

if diff -q /etc/presto/config.properties /etc/presto/config.properties.new &>/dev/null; then
  echo "Configuring Presto Server at the Coordinator"
  mv /etc/presto/config.properties.new /etc/presto/config.properties
  restart=true
else
  rm -f /etc/presto/config.properties.new
fi

if [[ "${restart}" == "true" ]]; then
  echo "Restarting Presto Server at the Coordinator"
  /etc/init.d/presto restart
fi
