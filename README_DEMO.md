## Demo scenario:

1. Show slides
  - https://docs.google.com/presentation/d/1A4b2J1HP1-aaGtYAVBXi5spkpbwB7eZdkcXz9Dk2Lzc/edit?usp=sharing

2. Init demo. Frank & John run prod
  - `./tools/demo-init.sh` to load all objects into Aptomi
  - `aptomi policy apply --show` to deploy initial state
  - Initial state is
    - Prod is in cluster-us-east
      - Frank (Ops) shares prod instance of twitter_stats with the org
      - John (Ops) shares prod instance of analytics_pipeline with org
  - Show that it got deployed on k8s
    - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
  - Run aptomi again
    - `aptomi policy apply --noop` - to ensure there are no more changes to apply
  - Show endpoints
    - `aptomi endpoint show`
  - Open Tweeviz UI
    - Shows SF, NY, Boston tweets

3. Alice & Bob deploy stage
  - Alice (ID=100) deploys new staging version of TS viz (Canary testing/updates)
  - Bob (ID=101) deploys new staging version of TS (Mexico tweets)
  - Run aptomi
    - `aptomi policy add dependencies demo/dependencies/dependencies.alice-stage-ts.yaml`
    - `aptomi policy add dependencies demo/dependencies/dependencies.bob-stage-ts.yaml`
    - `aptomi policy apply --noop --show` (explain share and reuse of services)
    - `aptomi policy apply`
  - Explain
    - Stage is in cluster-us-west
    - Alice & Bob have their own twitter apps
    - Alice & Bob share analytics pipeline
  - Show that it got deployed on k8s
    - `watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods`
  - Run aptomi again
    - `aptomi policy apply --noop` - to ensure there are no more changes to apply
  - Show endpoints
    - `aptomi endpoint show`
  - Open Tweeviz UI
    - Show stage Alice (different UI)
    - Show prod Bob (Mexico tweets)

4. Alice deletes her staging instance and asks Frank to propagate her new VS to production
  - Run aptomi
    - `aptomi policy delete dependencies demo/dependencies/dependencies.alice-stage-ts.yaml`
    - `vim dependencies.frank-prod-ts.yaml` and change demo-v51 -> demo-v52
    - `aptomi policy add dependencies demo/dependencies/dependencies.frank-prod-ts.yaml`
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


TODO:
- istio allocation in different clusters