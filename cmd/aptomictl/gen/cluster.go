package gen

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin/helm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func newClusterCommand(cfg *config.Client) *cobra.Command {
	var sourceContext, clusterName, clusterNamespace string
	var local bool

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "gen cluster",
		Long:  "gen cluster long",

		Run: func(cmd *cobra.Command, args []string) {
			if !local && len(sourceContext) == 0 {
				panic(fmt.Sprintf("context should be specified"))
			}
			if local && len(sourceContext) > 0 {
				panic(fmt.Sprintf("one of local or context could be specified"))
			}

			var clusterConfig *helm.Config
			var err error

			if local {
				if len(clusterName) == 0 {
					clusterName = "local"
				}
				clusterConfig = &helm.Config{Local: true, Namespace: "default"}
			} else {
				if len(clusterName) == 0 {
					clusterName = sourceContext
				}
				clusterConfig, err = handleKubeConfigCluster(sourceContext)
			}
			if err != nil {
				panic(err)
			}

			cluster := lang.Cluster{
				TypeKind: lang.ClusterObject.GetTypeKind(),
				Metadata: lang.Metadata{
					Name:      clusterName,
					Namespace: clusterNamespace,
				},
				Type:   "kubernetes",
				Config: clusterConfig,
			}

			data, err := yaml.Marshal(cluster)
			if err != nil {
				panic(fmt.Sprintf("error while marshaling generated cluster: %s", err))
			}

			fmt.Println(string(data))
		},
	}

	cmd.Flags().BoolVarP(&local, "local", "l", false, "Build Aptomi cluster with local kubernetes")
	cmd.Flags().StringVarP(&sourceContext, "context", "c", "", "Context in kubeconfig to be used for Aptomi cluster creation")
	cmd.Flags().StringVarP(&clusterName, "name", "n", "", "Name of the Aptomi cluster to create")
	cmd.Flags().StringVarP(&clusterNamespace, "namespace", "N", "system", "Namespace of the Aptomi cluster to create")

	return cmd
}

func handleKubeConfigCluster(sourceContext string) (*helm.Config, error) {
	kubeConfig, err := buildTempKubeConfigWith(sourceContext)
	if err != nil {
		return nil, fmt.Errorf("error while building temp kube config with context %s: %s", sourceContext, err)
	}

	clusterConfig := helm.Config{
		KubeConfig: kubeConfig,
	}

	return &clusterConfig, err
}

func buildTempKubeConfigWith(sourceContext string) (*interface{}, error) {
	rawConf, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	newConfig := api.NewConfig()
	newConfig.CurrentContext = sourceContext

	context, exist := rawConf.Contexts[sourceContext]
	if !exist {
		return nil, fmt.Errorf("requested context not found: %s", sourceContext)
	}
	newConfig.Contexts[sourceContext] = context

	if newConfig.Clusters[context.Cluster], exist = rawConf.Clusters[context.Cluster]; !exist {
		return nil, fmt.Errorf("requested cluster (from specified context) not found: %s", context.Cluster)
	}

	if newConfig.AuthInfos[context.AuthInfo], exist = rawConf.AuthInfos[context.AuthInfo]; !exist {
		return nil, fmt.Errorf("requested auth info (user from specified context) not found: %s", context.AuthInfo)
	}

	content, err := clientcmd.Write(*newConfig)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling temp kubeconfig: %s", err)
	}

	kubeConfig := new(interface{})
	err = yaml.Unmarshal(content, kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling temp kubeconfig: %s", err)
	}

	return kubeConfig, err
}

func getKubeConfig() (*api.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	conf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	rawConf, err := conf.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error while getting raw kube config: %s", err)
	}

	return &rawConf, err
}
