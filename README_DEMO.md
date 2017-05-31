## Demo scenario:

1. Show slides
   - https://docs.google.com/presentation/d/1A4b2J1HP1-aaGtYAVBXi5spkpbwB7eZdkcXz9Dk2Lzc/edit?usp=sharing

2. Show policy
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
   - Explain what will get matched
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
     - `kubectl --context kapp-demo -n demo get pods`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Shows SF, NY, Boston tweets

4. Deploy AP + TS for user Bob (ID=2)
   - Explain what will get matched
     - low-Bob (priority < 200), team-platform-services (priority < 200)
     - meaning, dedicated TS and shared AP
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Shows Japan tweets

5. Alice (ID=1) deploys new staging version of TS in parallel
   - With "demo-v42" and "stage" ts tags
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply`
   - Open Tweeviz UI
     - Shows SF, NY, Boston tweets. But different UI

6. Alice (ID=1) propagates staging version to production
   - Staging TS gets deleted
   - Production TS gets updated
   - Make sure to use `./aptomi policy apply --noop --show --verbose` to see deletions and updates
   - Refresh in browser
     - Stage instance disappears
     - Prod instance changes look and feel to demo-v42

7. Change # of top tweets
   - Default -> 3 and redeploy

7. Alice (ID=1) gets marked as "compromised"
   - Loses access to her "prod"
   - Right now all objects get deleted, but this behavior will be customizable

8. Low-priority user Carol (ID=3)
   - Gets nothing due to priority < 50

9. Deploy dedicated DAP for Carol (ID=3) in its own k8s cluster
   - Priority = 200
   - Show kubectl output
     - `kubectl --context kapp-demo -n demo get pods`
     - `kubectl --context minikube -n demo get pods`
