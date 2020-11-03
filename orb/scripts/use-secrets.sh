use_secrets() {
    cat "$FILE" >> $BASH_ENV
}

# Will not run if sourced for bats.
# View src/tests for more information.
TEST_ENV="bats-core"
if [ "${0#*$TEST_ENV}" == "$0" ]; then
    use_secrets
fi