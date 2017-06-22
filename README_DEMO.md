## Demo slides:
    - https://docs.google.com/a/renski.com/presentation/d/10X8oEjBqLPSfxf5KHhg8n_tBtYa7bAERHkZZiynjwyc/edit?usp=sharing

## Demo steps

### Preparation

1. Switch to the Aptomi GCloud account and project

For Sergey:
```shell
gcloud config set account sergei@apptomi.com
gcloud config set project symmetric-fin-171202
```

For Roman:
```shell
gcloud config set account roman@apptomi.com
gcloud config set project bright-torus-169502
```

2. Create k8s clusters
    - `./tools/demo-gke.sh up`

3. Initialize demo repository on GitHub and start auto-apply
    - ./tools/demo-init.sh
    - ./tools/demo-push.sh
    - ./tools/demo-watch-apply.sh

### Demo scenario

1. Explain initial state
    - No services have been defined
    - Aptomi only knows about 2 kubernetes clusters (us-east and us-west) and 5 users

2. Frank comes and defines analytics-pipeline


## Suntrust demo / story:

2. Show policy
   - We have a language.
      - Ops guy defined a policy - service & context
      - Context = who it's for and how resources are allocated/shared
   - Show k8s clusters
      - `kubectl config get-contexts`
      - `kubectl config view`
   - Show services
      - Analytics Pipeline
      - Twitter Stats
   - Show users
   - Show dependencies
   - Show contexts (where the secret sauce is)
      - Twitter Stats
      - Analytics Pipeline

3. Deploy AP + TS for user Alice (ID=1)
   - Explain what will get matched (Alice doesn't specify, policy controls that)
     - low-Alice (priority < 200), team-platform-services (priority < 200)
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply`
   - Run aptomi again
     - `./aptomi policy apply` - to ensure there are no more changes to apply
   - While it's loading, we can show tracing
     - `./aptomi policy apply --noop --trace`
   - Show kubectl output
     - `kubectl --context cluster-us-west -n demo get pods`
     - `watch -n1 -d -- kubectl --context cluster-us-rwest -n demo get pods`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Shows SF, NY, Boston tweets

4. Deploy AP + TS for user Bob (ID=2). Alice (ID=1) deploys new staging version of TS (Canary testing/updates)
  - Alice - with "demo-v42" and "stage" as tags
  - Explain what will get matched
     - low-Bob (priority < 200), team-platform-services (priority < 200)
     - meaning, dedicated TS and shared AP
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show` (explain share and reuse of services)
     - `./aptomi policy apply`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Show stage Alice (different UI)
     - Show prod Bob (Japan tweets)

5. Alice (ID=1) propagates staging version to production (making a change)
   - Staging TS gets deleted
   - Production TS gets updated
   - Make sure to use `./aptomi policy apply --noop --show --verbose` to see deletions and updates
   - Refresh in browser
     - Stage instance disappears
     - Prod instance changes look and feel to demo-v42

6. Change # of top tweets
   - Default -> 3 and redeploy

7. Alice (ID=1) gets marked as "compromised"
   - Loses access to her "prod"
   - Right now all objects get deleted, but this behavior will be customizable

8. Low-priority user Carol (ID=3)
   - Gets nothing due to priority < 50

9. Deploy dedicated DAP for Carol (ID=3) in its own k8s cluster
   - Priority = 200
   - Show kubectl output
     - `kubectl --context cluster-us-east -n demo get pods`
     - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
