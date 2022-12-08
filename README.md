<p align="center"><img width="150px" height="150px" src="logo.png" /></p>

<h1 align="center"> aca-cli - CLI helper tool to deploy applications to Azure Container Apps  </h1>

<p> This tool allows to deploy to multiple environments using one yaml definition file.  </p>

<h2> Example </h2>

When `aca-cli deploy` command is run it will look for `deploy.yaml` file in the repository root and parses it for instructions what to deploy. Here is example file 

```yaml
environments:
  - name: dev
    subscription_id: 4b549055-e193-4bc7-8e81-1234567
    resource_group: container-app-test
    container_app_name: products-api
    container_app_environment: container-app-test
    location: westeurope
    target_port: 8080
    containers:
      - name: postgres
        image: postgres:latest
        env:
          - name: POSTGRES_USER
            value: postgres
          - name: POSTGRES_PASSWORD
            value: postgres
          - name: POSTGRES_DB
            value: postgres
      - name: products-api
        image: joonvena/product-api:latest
        env:
          - name: DATABASE_URL
            value: postgres://postgres:postgres@localhost:5432/postgres
          - name: ALLOW_ORIGINS
            value: "*"
```

This example defines single environment called dev. To deploy this using `aca-cli`:

```shell
aca-cli deploy -e dev
```

After the deployment is done it will output the URL for the application if Ingress is enabled. 

<h2> Review Environments </h2>

As we can define multi container applications it is possible to easily create short lived environments that live only as long as the pull request is open. This is possible with `aca-cli` by defining environment called `review` and target `aca-cli deploy -e review` and the aca-cli will handle the rest. When PR is merged `aca-cli delete -e review` command is run and the deployment gets cleared. 
