# AWS Elastic Container Registry

If you are running within AWS, and making use of a private [Elastic Container Registry (ECR)](https://aws.amazon.com/ecr/) to host your application images, then the Trivy plugin will be unable to scan those images unless access is granted to the registry through an [Identity and Access Managemnent (IAM)](https://aws.amazon.com/iam/) role assigned to the service account running the Trivy plugins.

Once an IAM role granting grant access to the ECR has been created, this can be assigned to the service account by including the following additional parameter when running the `helm upgrade --install` command.

```shell
--set scan.plugins.annotations.eks\\.amazonaws\\.com/role-arn=arn:aws:iam::<AWS_ACCOUNT_ID>:role/<ROLE_NAME>
```
where `<AWS_ACCOUNT_ID>` should be replaced witth your AWS account ID, and `<ROLE_NAME>` should be replaced with the name of the role granting access to the ECR.

This will now allow the Trivy plugin to scan your internal images for vulnerabilities.
