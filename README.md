# batflow
A batch workflow.

# Development

Follow the [guide](https://docs.temporal.io/application-development/foundations#run-a-development-server) to installing development environment dependencies.

Once the dependency preparation is complete, the repeat run can also execute `make server`. This will run a temporal dev server.

Then start the worker on the other terminal `make server`.

Finally, execute `make starter` to run the workflow.