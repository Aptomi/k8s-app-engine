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
   - Show contexts (where the secret sauce is)

3. Deploy AP + TS for user Alice (ID=1)
   - Explain what will get matched
     - low-Alice (priority < 200), team-platform-services (priority < 200)
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply --debug`
   - Run aptomi again
     - `./aptomi policy apply --debug` - to ensure there are no more changes to apply
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
   - While it's loading, we can show tracing
     - `./aptomi policy apply --noop --tracing`

4. Deploy AP + TS for user Bob (ID=2)
   - Explain what will get matched
     - low-Bob (priority < 200), team-platform-services (priority < 200)
     - meaning, dedicated TS and shared AP
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply --debug`
   - Run aptomi again
     - `./aptomi policy apply --debug` - to ensure there are no more changes to apply
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
   - While it's loading, we can show tracing
     - `./aptomi policy apply --noop --tracing`

5. Alice (ID=1) deploys new staging version of TS in parallel
   - Show TS
     - tweepub: different region
     - tweeviz: different HTML to visualize data

6. Alice (ID=1) propagates staging version to production
   - Staging TS gets deleted
   - Production TS gets updated

7. Alice (ID=1) gets marked as "untrusted"
   - Loses access to her "prod"

8. Low-priority user Carol (ID=3)
   - Gets nothing due to low priority

9. Increase priority for Carol (ID=3) and deploy dedicated DAP for her
   - In another k8s cluster


tracing
change tag
