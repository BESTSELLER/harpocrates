vault_login() {
    VERSION=1.2.3
    wget -O ./vault_$VERSION\_linux_amd64.zip https://releases.hashicorp.com/vault/$VERSION/vault_$VERSION\_linux_amd64.zip
    unzip -o vault_$VERSION\_linux_amd64.zip
    chmod +x vault
    mv vault /usr/bin/.
    vault login -method=userpass username=$VAULT_USERNAME password=$VAULT_PASSWORD >/dev/null
    echo 'export VAULT_TOKEN=$(cat $HOME/.vault-token)' >> $BASH_ENV
}



# Will not run if sourced for bats.
# View src/tests for more information.
TEST_ENV="bats-core"
if [ "${0#*$TEST_ENV}" == "$0" ]; then
    vault_login
fi