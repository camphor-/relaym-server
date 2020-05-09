#!/bin/bash

# マイグレーション
TMPDIR="/tmp"
TAR_FILE="/tmp/skeema.tar.gz"
RELEASES_URL="https://github.com/skeema/skeema/releases"
VERSION=1.4.2
curl -s -L -o "${TAR_FILE}" "${RELEASES_URL}/download/v${VERSION}/skeema_${VERSION}_linux_amd64.tar.gz"
tar -xf "${TAR_FILE}" -C "${TMPDIR}"
chmod +x "${TMPDIR}/skeema"
mv "${TMPDIR}/skeema" /usr/local/bin/skeema

cd /mysql
sed -i "s/schema=relaym/schema=relaym_${ENV}/" ./schemas/relaym/.skeema
skeema push -p"${DB_PASSWORD}" "${ENV}"

exec "$@"
