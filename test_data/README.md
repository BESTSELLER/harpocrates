# Secret file tests

This directory contains yaml files that are used to test the secret file json schema.

The secret file test is build to find all files in this directory and test them against the secret file json schema.
There's a few very important things to follow when creating tests for secret files:

- Files <u>must</u> be of extension `.yaml` or `.yml`
- Files that are expected to fail <u>must</u> include `not_valid` in the naming of the file
- Files that are expected to succeed <u>must not</u> include `not_valid` in the naming of the file
