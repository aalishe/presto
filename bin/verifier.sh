#!/usr/bin/env bash

if [[ "${REEVE_TAG_PRESTO_ROLE}" != "coordinator" ]]; then
  echo "$0 is only to be executed at the Reeve Coordinator." >&2
  exit 1
fi

SQLUSER='prestoverifier'
SQLPASSWORD='P@55word'

# Setup the verifier database and user
serf event mysql-presto-adapter/setverifier

if [[ ! -x /root/bin/verifier ]]; then
  mkdir -p /root/bin
  curl -o /root/bin/verifier https://repo1.maven.org/maven2/com/facebook/presto/presto-verifier/0.161/presto-verifier-0.161-executable.jar
  chmod +x /root/bin/verifier
fi

mysql_master_ip=$(serf members | grep 'presto_adapter=mysql' | grep 'mysql_role=master' | awk '{print $2}' | cut -f1 -d:)

if [[ ! -e /tmp/verifier.config.properties ]]; then
cat <<EOCONF >/tmp/verifier.config.properties
suite=my_suite
query-database=jdbc:mysql://${mysql_master_ip}:3306/presto_verifier?user=${SQLUSER}&password=${SQLPASSWORD}
control.gateway=jdbc:presto://localhost:8080
test.gateway=jdbc:presto://localhost:8081
thread-count=1
EOCONF
fi

/root/bin/verifier /tmp/verifier.config.properties

# If it is required to remove the verifier database and user execute:
# serf event mysql-presto-adapter/delverifier
