<p align="center"><img width="150px" height="150px" src="logo.png" /></p>

<h1 align="center"> aca-cli - CLI helper tool to deploy applications to Azure Container Apps  </h1>

<p> This tool allows to deploy to multiple environments using one yaml definition file. You can find demo repository that shows the usage with Github workflows from here: <a href="https://github.com/joonvena/aca-cli-demo">https://github.com/joonvena/aca-cli-demo</a> </p>

<h2> Example </h2>

When `aca-cli deploy` command is run it will look for `deploy.yaml` file in the repository root and parses it for instructions what to deploy. Here is example file 

```yaml
environments:
  - name: dev
    subscription_id: 4b549055-e193-4bc7-8e81-1234567
    resource_group: container-app-test
    container_app_name: aca-cli-demo
    container_app_environment: container-app-test
    location: westeurope
    containers:
      - name: aca-cli-demo
          image: joonvena/aca-cli-demo:$tag
          env:
            - name: DATABASE_URL
              value: postgres://postgres:postgres@localhost:5432/postgres
            - name: ALLOW_ORIGINS
              value: "*"
      - name: postgres
        image: postgres:latest
        env:
          - name: POSTGRES_USER
            value: postgres
          - name: POSTGRES_PASSWORD
            value: postgres
          - name: POSTGRES_DB
            value: postgres
```

For now the main application container is expected to be defined as first container in `containers` block. Purpose of the `$tag` is to tell the `aca-cli` that this value should be replaced. Currently it looks value for this in `GITHUB_SHA` variable or you can also give the tag from the cli using `-t` flag.

By default Ingress is set as external and port is set to 8080. If you want to override this you can add following definition to the environment:

```yaml
ingress:
  external: false
  targetPort: <port>
```

If you want to set secrets you can use `secretRef` in `env` block:

```yaml
env:
  ...
  - name: SECRET
    secretRef: secret-value
```

When application is deployed it will look `SECRET` variable from env, create new secret with name `secret-value` and injects the value of variable to this secret.

To deploy this:

```shell
aca-cli deploy -e dev
```

After the deployment is done it will output the URL for the application if Ingress is set as External.

<h3> Authentication </h3>

This uses the DefaultAzureCredential implementation for the authentication so you have many options for authentication. When running in `CI` you can use `EnvironmentCredential` by creating SP in Azure and defining following environment variables:

- AZURE_CLIENT_ID
- AZURE_TENANT_ID
- AZURE_CLIENT_SECRET

When running locally you could utilize the `az cli` credentials.

<h3> Private registry </h3>

If you need to get images from private registry you need to define `registries` block in the environment config.

```yaml
registries:
  - server: ghcr.io
    username: username
    PasswordSecretRef: ghcr-io-registry-credentials
```

When deployment is done it will check if there are any private registries defined and will create the secret using the `passwordSecretRef` as name of the secret. Value of the secret will be sourced from environment variable. In this case you need to have environment variable defined matching `GHCR_IO_REGISTRY_PASSWORD` otherwise the secret will have empty value.

<h2> Review Environments </h2>

As we can define multi container applications it is possible to easily create short lived environments that live only as long as the pull request is open. This is possible with `aca-cli` by defining environment called `review` and target `aca-cli deploy -e review -a <app_name>` . When PR is merged `aca-cli delete -e review -a <app_name>` command can be run to get the environment cleaned up. Purpose of the `-a` flag is to define unique name for the container app so it gives flexibility for the user to define their own patterns how to handle the naming.

