FROM scratch
ENTRYPOINT ["/kube-cron-rollout-restart"]
COPY kube-cron-rollout-restart /