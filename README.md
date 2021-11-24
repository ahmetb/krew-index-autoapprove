# krew-autoapprove

Automatically approves straightforward version updates to [krew-index]
repository.

[krew-index]: https://github.com/kubernetes-sigs/krew-index

## Deploy

Re-deploy via:

```sh
gcloud run deploy krew-autoapprove --project ahmet-personal-api \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --image $(KO_DOCKER_REPO=gcr.io/ahmet-personal-api/krew-autoapprove/webhook ko publish ./webhook)
```

First-time deploys: also add `--set-env-vars=GITHUB_TOKEN=...`
