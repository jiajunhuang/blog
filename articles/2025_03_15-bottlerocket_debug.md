# BottleRocket Linux check kubelet logs

I'm trying to setup a nodegroup in EKS, but it fails with error log:
"NodeCreationFailure Instances failed to join the kubernetes cluster", the system log of EC2 instance shows
"[FAILED] Failed to start Kubelet. See 'systemctl status kubelet.service' for details".

So I need to dig into BottleRocket Linux, but it's a container based system, here is the steps:

1. Connect to BottleRocket Linux EC2 by using: `aws ssm start-session --target INSTANCE_ID --region REGION_CODE`, then
you will be attached to the default control container.
2. Access the admin container from  default control container by execute `enter-admin-container`
3. Execute `apiclient exec admin bash`
4. Execute `sheltie` to gain the full root shell in the BottleRocket Linux host
5. And now you can execute commands like `systemctl/journal` and so on.

I see a error in the kubelet log:

```bash
failed to validate kubelet flags: unknown 'kubernetes.io' or 'k8s.io' labels specified with --node-labels: [app.kubernetes.io/component app.kubernetes.io/instance app.kubernetes.io/name]
```

Which indicates that the reason why the kubelet fails to start is I use a forbidden labels in my node, double check it:

```bash
$ apiclient get settings.kubernetes.node-labels
{
  "settings": {
    "kubernetes": {
      "node-labels": {
        "app.kubernetes.io/component": "my-app",
        "app.kubernetes.io/instance": "my-instance",
        "app.kubernetes.io/name": "my-app",
        "eks.amazonaws.com/capacityType": "ON_DEMAND",
        "eks.amazonaws.com/nodegroup": "ng-87ba2ae2838",
        "eks.amazonaws.com/nodegroup-image": "ami-07e8bf04276e4fd60",
        "eks.amazonaws.com/sourceLaunchTemplateId": "lt-020692bf92cf90040",
        "eks.amazonaws.com/sourceLaunchTemplateVersion": "1"
      }
    }
  }
}
```

After I remove the labels in my node, and run it again, it works :)
