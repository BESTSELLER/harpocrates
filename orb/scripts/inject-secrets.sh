inject_secrets() {
    if [ -f ${BASH_ENV} ]; then
        source ${BASH_ENV}
    fi

    if [ -f "/tmp/secrets" ]; then
        source /tmp/secrets
    fi

    if [ -z "$CONTAINER_NAME"]; then
        export CONTAINER_NAME=$APP_NAME
    fi

    export SECRETS=$(yq read $SECRET_FILE -j)

    curl -s -H "Accept:application/vnd.github.v3.raw" -o $DEPLOYMENT_TYPE.yml -L https://github.com/BESTSELLER/harpocrates/releases/download/$HARPOCRATES_VERSION/$DEPLOYMENT_TYPE.yml
    curl -s -H "Accept:application/vnd.github.v3.raw" -o kustomization.yml -L https://github.com/BESTSELLER/harpocrates/releases/download/$HARPOCRATES_VERSION/kustomization.yml

    envsubst < ./kustomization.yml > ./kustomization_var.yml
    mv ./kustomization_var.yml ./kustomization.yml

    envsubst < ./$DEPLOYMENT_TYPE.yml > ./$DEPLOYMENT_TYPE_var.yml
    mv ./$DEPLOYMENT_TYPE_var.yml ./$DEPLOYMENT_TYPE.yml

    kubectl kustomize . > new.yml
    mv new.yml $DEPLOY_FILE
}

# Will not run if sourced for bats.
# View src/tests for more information.
TEST_ENV="bats-core"
if [ "${0#*$TEST_ENV}" == "$0" ]; then
    inject_secrets
fi