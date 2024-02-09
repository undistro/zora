# HTTPS Proxy

If your network environment requires the use of a proxy, you must ensure proper configuration of the `httpsProxy`
parameter when running `helm upgrade --install` command.

```shell 
# omitted "helm upgrade --install" command and parameters

--set httpsProxy="https://secure.proxy.tld"
```

Additionally, you can specify URLs that should bypass the proxy, by setting the `noProxy` parameter in comma-separated 
list format. Note that this parameter already has a default value: `kubernetes.default.svc.*,127.0.0.1,localhost`.

Configuring proxy settings enables both `trivy` plugin and `zora-operator` to use the proxy for external requests.

Zora OSS sends scan results to the following external URL if your installation is integrated with
[Zora Dashboard](../dashboard.md):

- `https://zora-dashboard.undistro.io`

While [Trivy](../plugins/trivy.md) downloads vulnerability databases during scans from the following external sources:

- `ghcr.io/aquasecurity/trivy-db` 
- `ghcr.io/aquasecurity/trivy-java-db`
