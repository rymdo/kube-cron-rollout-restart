# kube-cron-rollout-restart

The kube-cron-rollout-restart is used to restart workloads in a kubernetes cluster base on CRON Schedules set with annotations. Useful for eg. development environments that need to run latest version of containers (targeting `latest` or a mutable tag).

## Usage

### Launch Parameters

```
  --kubeconfig
        use kubeconfig (default is in-cluster)
  --kubeconfig-path
        path to kubeconfig (default "<HOME>/.kube/config")
  --alertmanager-url
        url to alertmanager (default "http://alertmanager:80")
```

### Annotations

Use these annotations on the workload

| Name                                                 | Value             | Description                                                                                                                      |
| ---------------------------------------------------- | ----------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `cron.rollout.restart/schedule`                      | `<cron-schedule>` | Full CRON Schedule. Eg. `"0 12 * * 1-5"` will restart the workload every weekday at 12:00 (`https://crontab.guru/#0_12_*_*_1-5`) |
| `cron.rollout.restart/alertmanager-silence-enabled`  | `true`\|`false`   | Enable/Disable Alertmanager Silence for workload. Default is `"false"`                                                           |
| `cron.rollout.restart/alertmanager-silence-duration` | `<duration>`      | Duration in minutes. (Default `"15"`)                                                                                            |
| `cron.rollout.restart/alertmanager-silence-labels`   | `<labels>`        | Comma separated silence matching labels, eg. `"key1=value1,key2=value2"`                                                         |
| `cron.rollout.restart/alertmanager-silence-comment`  | `<comment>`       | Comment for the silence. (Default `"kube-cron-rollout-restart"`)                                                                 |

### RBAC

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-cron-rollout-restart
rules:
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets"]
  verbs: ["list", "get", "patch"]
```

## Links

Parts of alertmanager code got from https://github.com/snigdhasambitak/alertmanager-silence-cli
