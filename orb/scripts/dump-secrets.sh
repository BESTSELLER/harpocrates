dump_secrets() {
    mkdir -p $OUTPUT
        
    # This is the default one, if no vault-path is defined
    if [ "$VAULT_PATH" == "" ] ; then

        # Fetch all the "common" secrets if that env var has been set
        if [ "$HARPOCRATES_SECRETS" != "" ] ; then
            /harpocrates \
                --format "$FORMAT" \
                --output "$OUTPUT" \
                --prefix "K8S_CLUSTER_" \
                --secret $HARPOCRATES_SECRETS
        fi

        # Fetch cluster credentials
        if [ "$CLUSTER_SECRET" != "" ] ; then
            export HARPOCRATES_FILENAME="cluster_secret.json"
            /harpocrates \
                --format json \
                --output "$OUTPUT" \
                --secret "$CLUSTER_SECRET"
            unset HARPOCRATES_FILENAME
        fi

        # Fetch docker credentials
        if [ "$SHORT" != "" ] ; then
            /harpocrates \
                --format "$FORMAT" \
                --output "$OUTPUT" \
                --prefix "DOCKER_" \
                --secret "ES/data/service_accounts/harbor/$SHORT-ci"
        fi

    fi

    # Check if they have specified vault-path and then fetch those from Vault
    if [ "$VAULT_PATH" != "" ] ; then
        if [ "$OUTPUT_FILENAME" != "" ] ; then
            export HARPOCRATES_FILENAME="$OUTPUT_FILENAME"
        fi

        /harpocrates \
            --format "$FORMAT" \
            --output "$OUTPUT" \
            --secret "$VAULT_PATH"
    fi
}

# Will not run if sourced for bats.
# View src/tests for more information.
TEST_ENV="bats-core"
if [ "${0#*$TEST_ENV}" == "$0" ]; then
    dump_secrets
fi