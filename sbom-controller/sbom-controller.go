package main

import (
    "context"
    "fmt"
    "os/exec"
    "time"
    "io/ioutil"
    "path/filepath"

    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        panic(err)
    }

    dynamicClient, err := dynamic.NewForConfig(config)
    if err != nil {
        panic(err)
    }

    sbomGVR := schema.GroupVersionResource{
        Group:    "example.com",
        Version:  "v1",
        Resource: "sboms",
    }

    for {
        sboms, err := dynamicClient.Resource(sbomGVR).Namespace("default").List(context.TODO(), metav1.ListOptions{})
        if err != nil {
            panic(err)
        }

        for _, sbom := range sboms.Items {
            image, found, err := unstructured.NestedString(sbom.Object, "spec", "image")
            if err != nil || !found {
                continue
            }

            sbomData, err := generateSBOM(image)
            if err != nil {
                continue
            }

            unstructured.SetNestedField(sbom.Object, sbomData, "spec", "sbom")
            dynamicClient.Resource(sbomGVR).Namespace("default").Update(context.TODO(), &sbom, metav1.UpdateOptions{})

            err = commitSBOMToGit(sbom.GetName(), sbomData)
            if err != nil {
                fmt.Printf("Failed to commit SBOM to Git: %v\n", err)
            }
        }

        time.Sleep(30 * time.Second)
    }
}

func generateSBOM(image string) (string, error) {
    cmd := exec.Command("syft", image, "-o", "json")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}

func commitSBOMToGit(sbomName, sbomData string) error {
    repoPath := "/path/to/your/repo"
    sbomFilePath := filepath.Join(repoPath, "sboms", sbomName+".json")

    err := ioutil.WriteFile(sbomFilePath, []byte(sbomData), 0644)
    if err != nil {
        return err
    }

    cmd := exec.Command("git", "-C", repoPath, "add", sbomFilePath)
    if err := cmd.Run(); err != nil {
        return err
    }

    cmd = exec.Command("git", "-C", repoPath, "commit", "-m", fmt.Sprintf("Add SBOM for %s", sbomName))
    if err := cmd.Run(); err != nil {
        return err
    }

    cmd = exec.Command("git", "-C", repoPath, "push")
    if err := cmd.Run(); err != nil {
        return err
    }

    return nil
}
