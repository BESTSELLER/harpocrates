setup() {
     # load the functions we need
    source ./scripts/use-secrets.sh

    # make sure bash_env is set.. just in case..
    if [ -z "$BASH_ENV" ]; then
        export BASH_ENV=/tmp/bash_env
    fi

    # initial test data
    export FILE="/tmp/secrets"
    echo "export SECRET_HERE=123" > $FILE
}

teardown() {
    rm /tmp/secrets
    rm /tmp/bash_env
}

@test "1: Secrets are added to BASH_ENV" {
    [ "$SECRET_HERE" == "" ]
    use_secrets
    source $BASH_ENV
    [ "$SECRET_HERE" == "123" ]
}