package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	getRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/get"
	openshiftRootCmd "github.com/jkandasa/autoeasy/cmd/plugin/openshift/root"
	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
	"github.com/mycontroller-org/server/v2/pkg/utils/printer"
	csvv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/pkg/api"
	"github.com/operator-framework/operator-registry/pkg/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type RelatedImage struct {
	PackageName string `json:"packageName" yaml:"packageName"`
	Channel     string `json:"channel" yaml:"channel"`
	CsvName     string `json:"csvName" yaml:"csvName"`
	ImageName   string `json:"imageName" yaml:"imageName"`
	Image       string `json:"image" yaml:"image"`
}

const (
	IndexImageNamespace     = "autoeasy-index-image-ns"
	IndexImagePodName       = "index-image"
	IndexImageContainerPort = 50051
)

var (
	indexImage           string
	registryAddress      string
	listAllPackages      bool
	displayRelatedImages bool
)

func init() {
	getRootCmd.AddCommand(getPackageCmd)
	getPackageCmd.Flags().StringVar(&indexImage, "image", openshiftRootCmd.DefaultRegistryImage, "index image name")
	getPackageCmd.Flags().StringVar(&registryAddress, "address", "", "registry address. if left blank, index image will be deployed on your cluster by this tool")
	getPackageCmd.Flags().BoolVar(&listAllPackages, "all", false, "list all packages")
	getPackageCmd.Flags().BoolVar(&displayRelatedImages, "related-images", false, "displays only related images")
}

var getPackageCmd = &cobra.Command{
	Use:   "package",
	Short: "prints package details from a registry",
	Run: func(cmd *cobra.Command, packagesList []string) {
		if !listAllPackages && len(packagesList) == 0 {
			zap.L().Error("there is no package name supplied")
			rootCmd.ExitWithError()
		}
		printRegistryImages(registryAddress, packagesList)
	},
}

func printRegistryImages(address string, packagesList []string) {
	// deploy index image, if required
	closePortForwardFunc, address, err := deployIndexImage(address)
	if err != nil {
		zap.L().Error("error on deploying index image", zap.String("indexImage", indexImage), zap.Error(err))
		rootCmd.ExitWithError()
	}

	// close the port forward binding
	defer func() {
		if closePortForwardFunc != nil {
			closePortForwardFunc()
		}
		undeployIndexImage()
	}()

	// get operator registry client
	registryClient, err := client.NewClient(address)
	if err != nil {
		zap.L().Error("error on loading k8s client", zap.Error(err))
		rootCmd.ExitWithError()
	}
	ctx := context.Background()

	// displays related images and return
	if displayRelatedImages {
		printRegistryRelatesImages(ctx, registryClient, address, packagesList)
		return
	}

	headers := []printer.Header{
		{Title: "package name", ValuePath: "packageName"},
		{Title: "channel", ValuePath: "channelName"},
		{Title: "version", ValuePath: "version"},
		{Title: "csv name", ValuePath: "csvName"},
		{Title: "skip range", ValuePath: "skipRange"},
		{Title: "replaces", ValuePath: "replaces"},
		{Title: "bundle path", ValuePath: "bundlePath", IsWide: true},
	}

	rows := make([]interface{}, 0)

	if listAllPackages {
		bundles, err := registryClient.ListBundles(ctx)
		if err != nil {
			zap.L().Error("error on getting bundles", zap.Error(err))
			rootCmd.ExitWithError()
		}
		for {
			bundle := bundles.Next()
			if bundle == nil {
				break
			}
			rows = append(rows, bundle)
		}

	} else {
		for _, packageName := range packagesList {
			pkg, err := registryClient.GetPackage(ctx, packageName)
			if err != nil {
				zap.L().Error("error on getting packages", zap.Error(err))
				continue
			}

			for _, channel := range pkg.Channels {
				// get bundle
				bundle, err := registryClient.GetBundle(ctx, packageName, channel.Name, channel.CsvName)
				if err != nil {
					zap.L().Error("error on getting a bundle", zap.Error(err))
					continue
				}

				// update default channel
				if pkg.DefaultChannelName == channel.Name {
					bundle.ChannelName = fmt.Sprintf("%s [d]", bundle.ChannelName)
				}

				rows = append(rows, bundle)
			}

		}
	}
	printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}

func printRegistryRelatesImages(ctx context.Context, registryClient *client.Client, address string, packagesList []string) {
	headers := []printer.Header{
		{Title: "package name", ValuePath: "packageName"},
		{Title: "channel", ValuePath: "channel"},
		{Title: "csv name", ValuePath: "csvName"},
		{Title: "image name", ValuePath: "imageName"},
		{Title: "image", ValuePath: "image"},
	}

	rows := make([]interface{}, 0)

	if listAllPackages {
		bundles, err := registryClient.ListBundles(ctx)
		if err != nil {
			zap.L().Error("error on getting bundles", zap.Error(err))
			rootCmd.ExitWithError()
		}

		for {
			bundle := bundles.Next()
			if bundle == nil {
				break
			}
			updatedBundle, err := registryClient.GetBundle(ctx, bundle.PackageName, bundle.ChannelName, bundle.CsvName)
			if err != nil {
				zap.L().Error("error on getting updated bundle with csv json", zap.String("package", bundle.PackageName), zap.String("channel", bundle.ChannelName), zap.Error(err))
				continue
			}

			_rows, err := loadRelatedImages(updatedBundle)
			if err != nil {
				zap.L().Error("error on getting clusterServiceVersion details", zap.String("package", bundle.PackageName), zap.String("channel", bundle.ChannelName), zap.Error(err))
				continue
			}
			rows = append(rows, _rows...)
		}

	} else {
		for _, packageName := range packagesList {
			pkg, err := registryClient.GetPackage(ctx, packageName)
			if err != nil {
				zap.L().Error("error on getting packages", zap.Error(err))
				continue
			}

			for _, channel := range pkg.Channels {
				// get bundle
				bundle, err := registryClient.GetBundle(ctx, packageName, channel.Name, channel.CsvName)
				if err != nil {
					zap.L().Error("error on getting a bundle", zap.Error(err))
					continue
				}

				// update default channel
				if pkg.DefaultChannelName == channel.Name {
					bundle.ChannelName = fmt.Sprintf("%s [d]", bundle.ChannelName)
				}

				_rows, err := loadRelatedImages(bundle)
				if err != nil {
					zap.L().Error("error on getting clusterServiceVersion details", zap.String("package", bundle.PackageName), zap.String("channel", bundle.ChannelName), zap.Error(err))
					continue
				}
				rows = append(rows, _rows...)

			}

		}
	}
	printer.Print(os.Stdout, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}

func loadRelatedImages(bundle *api.Bundle) ([]interface{}, error) {
	csv := &csvv1alpha1.ClusterServiceVersion{}
	err := json.Unmarshal([]byte(bundle.GetCsvJson()), csv)
	if err != nil {
		return nil, err
	}
	rows := make([]interface{}, 0)

	for _, image := range csv.Spec.RelatedImages {
		row := RelatedImage{
			PackageName: bundle.PackageName,
			Channel:     bundle.ChannelName,
			CsvName:     bundle.CsvName,
			ImageName:   image.Name,
			Image:       image.Image,
		}

		rows = append(rows, row)
	}
	return rows, nil
}
