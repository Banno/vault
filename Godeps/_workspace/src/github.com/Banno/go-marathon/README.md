## Go-Marathon

This is a barebones, lightweight connector to the Marathon v2 API. It's initially meant to be used to control `apps` and `groups` via something like Terraform.

This is very much a work in progress.

## Testing

To run the acceptance tests against a locally running marathon you can do a `big up -d marathon` and then a `./test-local.sh`.
