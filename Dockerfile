FROM alpine:3.15
ENTRYPOINT ["/kube-cron-rollout-restart"]
COPY kube-cron-rollout-restart /