package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// AutopilotPoolConfig requests capacity by vCPU rather than a fixed server class.
type AutopilotPoolConfig struct {
	Name            string `json:"name"`
	VCPUs           int64  `json:"vcpus"`
	MemoryPerVCPU   string `json:"memoryPerVCPU"`
	BudgetPerHour   string `json:"budgetPerHour"`
	MinVCPUsPerNode int64  `json:"minVCPUsPerNode,omitempty"`
	MaxVCPUsPerNode int64  `json:"maxVCPUsPerNode,omitempty"`
}

var autopilotBudgetPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

func validateAutopilotPools(pools []AutopilotPoolConfig) error {
	names := map[string]bool{}
	for _, p := range pools {
		if uuid.Validate(p.Name) != nil || p.Name != strings.ToLower(p.Name) {
			return fmt.Errorf("Autopilot pool name must be a lowercase UUID")
		}
		if names[p.Name] {
			return fmt.Errorf("duplicate Autopilot pool %s", p.Name)
		}
		names[p.Name] = true
		if p.VCPUs <= 0 {
			return fmt.Errorf("Autopilot pool %s: vcpus must be positive", p.Name)
		}
		switch p.MemoryPerVCPU {
		case "any", "2", "4", "8":
		default:
			return fmt.Errorf("Autopilot pool %s: memoryPerVCPU must be any, 2, 4, or 8", p.Name)
		}
		budget, ok := new(big.Rat).SetString(p.BudgetPerHour)
		if !ok || budget.Sign() <= 0 || !autopilotBudgetPattern.MatchString(p.BudgetPerHour) {
			return fmt.Errorf("Autopilot pool %s: budgetPerHour must be a positive decimal", p.Name)
		}
		if p.MinVCPUsPerNode < 0 || p.MaxVCPUsPerNode < 0 || (p.MaxVCPUsPerNode > 0 && p.MinVCPUsPerNode > p.MaxVCPUsPerNode) {
			return fmt.Errorf("Autopilot pool %s: invalid per-node vCPU bounds", p.Name)
		}
	}
	return nil
}

type autopilotAPI interface {
	ListAutopilotPools(context.Context, string, string) ([]AutopilotPoolConfig, error)
	WriteAutopilotPool(context.Context, string, string, string, AutopilotPoolConfig, bool) error
	DeleteAutopilotPool(context.Context, string, string) error
}

// The upstream SDK v0.2.0 has no Autopilot resources. Reuse its configured
// transport, authentication and organization resolution for the console API.
type autopilotClient struct{ *spotv1.RackspaceSpotClient }

type autopilotResource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Labels    map[string]string `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		Region     string `json:"region"`
		CloudSpace string `json:"cloudSpace"`
		VCPU       struct {
			Total int64 `json:"total"`
		} `json:"vcpu"`
		MemoryPerVCPU      string `json:"memoryPerVCPU"`
		BudgetPerHour      string `json:"budgetPerHour"`
		AllocationStrategy string `json:"allocationStrategy"`
		VCPUPerNode        *struct {
			Min int64 `json:"min,omitempty"`
			Max int64 `json:"max,omitempty"`
		} `json:"vcpuPerNode,omitempty"`
	} `json:"spec"`
}

func (c *autopilotClient) request(ctx context.Context, method, path string, body any, out any) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.BaseURL, "/")+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	if method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/merge-patch+json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Rackspace Spot API request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Rackspace Spot API resource not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Rackspace Spot API returned HTTP %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(out)
	}
	return nil
}

func (c *autopilotClient) resourcePath(ctx context.Context, org, resource string) (string, string, error) {
	found, namespace, err := c.GetOrgID(ctx, org)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", fmt.Errorf("organization %q not found", org)
	}
	return "/apis/ngpc.rxt.io/v1/namespaces/" + url.PathEscape(namespace) + "/" + resource, namespace, nil
}

func (c *autopilotClient) ListAutopilotPools(ctx context.Context, org, cloudspace string) ([]AutopilotPoolConfig, error) {
	path, _, err := c.resourcePath(ctx, org, "autopilotnodepools")
	if err != nil {
		return nil, err
	}
	var list struct {
		Items []autopilotResource `json:"items"`
	}
	if err := c.request(ctx, http.MethodGet, path+"?labelSelector="+url.QueryEscape("ngpc.rxt.io/cloudspace="+cloudspace), nil, &list); err != nil {
		return nil, err
	}
	pools := make([]AutopilotPoolConfig, 0, len(list.Items))
	for _, r := range list.Items {
		p := AutopilotPoolConfig{Name: r.Metadata.Name, VCPUs: r.Spec.VCPU.Total, MemoryPerVCPU: r.Spec.MemoryPerVCPU, BudgetPerHour: r.Spec.BudgetPerHour}
		if r.Spec.VCPUPerNode != nil {
			p.MinVCPUsPerNode = r.Spec.VCPUPerNode.Min
			p.MaxVCPUsPerNode = r.Spec.VCPUPerNode.Max
		}
		pools = append(pools, p)
	}
	return pools, nil
}

func (c *autopilotClient) WriteAutopilotPool(ctx context.Context, org, cloudspace, region string, p AutopilotPoolConfig, exists bool) error {
	path, namespace, err := c.resourcePath(ctx, org, "autopilotnodepools")
	if err != nil {
		return err
	}
	r := autopilotResource{APIVersion: "ngpc.rxt.io/v1", Kind: "AutopilotNodePool"}
	r.Metadata.Name = p.Name
	r.Metadata.Namespace = namespace
	r.Metadata.Labels = map[string]string{"ngpc.rxt.io/cloudspace": cloudspace}
	r.Spec.Region = region
	r.Spec.CloudSpace = cloudspace
	r.Spec.VCPU.Total = p.VCPUs
	r.Spec.MemoryPerVCPU = p.MemoryPerVCPU
	r.Spec.BudgetPerHour = p.BudgetPerHour
	r.Spec.AllocationStrategy = "LowestPrice"
	if p.MinVCPUsPerNode != 0 || p.MaxVCPUsPerNode != 0 {
		r.Spec.VCPUPerNode = &struct {
			Min int64 `json:"min,omitempty"`
			Max int64 `json:"max,omitempty"`
		}{p.MinVCPUsPerNode, p.MaxVCPUsPerNode}
	}
	method := http.MethodPost
	var body any = r
	if exists {
		method = http.MethodPatch
		path += "/" + url.PathEscape(p.Name)
		// Explicit null clears previously configured bounds under JSON merge patch.
		patch := map[string]any{
			"region":             r.Spec.Region,
			"cloudSpace":         r.Spec.CloudSpace,
			"vcpu":               r.Spec.VCPU,
			"memoryPerVCPU":      r.Spec.MemoryPerVCPU,
			"budgetPerHour":      r.Spec.BudgetPerHour,
			"allocationStrategy": r.Spec.AllocationStrategy,
		}
		if r.Spec.VCPUPerNode == nil {
			patch["vcpuPerNode"] = nil
		} else {
			bounds := map[string]any{"min": nil, "max": nil}
			patch["vcpuPerNode"] = bounds
			if p.MinVCPUsPerNode > 0 {
				bounds["min"] = p.MinVCPUsPerNode
			}
			if p.MaxVCPUsPerNode > 0 {
				bounds["max"] = p.MaxVCPUsPerNode
			}
		}
		body = map[string]any{"spec": patch}
	}
	return c.request(ctx, method, path, body, nil)
}

func (c *autopilotClient) DeleteAutopilotPool(ctx context.Context, org, name string) error {
	path, _, err := c.resourcePath(ctx, org, "autopilotnodepools")
	if err != nil {
		return err
	}
	return c.request(ctx, http.MethodDelete, path+"/"+url.PathEscape(name), nil, nil)
}

// The console identifies generated child SpotNodePools using this label. Never
// reconcile them as manual pools or include them as editable pools on import.
func (c *autopilotClient) ListSpotNodePools(ctx context.Context, org, cloudspace string) ([]*spotv1.SpotNodePool, error) {
	path, _, err := c.resourcePath(ctx, org, "spotnodepools")
	if err != nil {
		return nil, err
	}
	var list spotv1.SpotNodePoolListResponse
	if err := c.request(ctx, http.MethodGet, path+"?labelSelector="+url.QueryEscape("ngpc.rxt.io/cloudspace="+cloudspace), nil, &list); err != nil {
		return nil, err
	}
	pools := make([]*spotv1.SpotNodePool, 0, len(list.Items))
	for _, r := range list.Items {
		if r.Metadata.Labels["autopilotnodepool.ngpc.rxt.io/name"] != "" {
			continue
		}
		p := &spotv1.SpotNodePool{Name: r.Metadata.Name, Org: org, Cloudspace: r.Spec.CloudSpace, ServerClass: r.Spec.ServerClass, Desired: r.Spec.Desired, BidPrice: r.Spec.BidPrice, WonCount: r.Status.WonCount, Status: r.Status.BidStatus, CustomLabels: r.Spec.CustomLabels, CustomAnnotations: r.Spec.CustomAnnotations, CustomTaints: r.Spec.CustomTaints}
		p.Autoscaling.Enabled = r.Spec.Autoscaling.Enabled
		p.Autoscaling.MinNodes = int64(r.Spec.Autoscaling.MinNodes)
		p.Autoscaling.MaxNodes = int64(r.Spec.Autoscaling.MaxNodes)
		pools = append(pools, p)
	}
	return pools, nil
}

func (c *autopilotClient) manualCloudspace(ctx context.Context, org, name string, original *spotv1.CloudSpace) (*spotv1.CloudSpace, error) {
	cs := *original
	pools, err := c.ListSpotNodePools(ctx, org, name)
	if err != nil {
		return nil, err
	}
	cs.SpotNodepools = nil
	for _, p := range pools {
		cs.SpotNodepools = append(cs.SpotNodepools, p)
	}
	return &cs, nil
}

func (c *spotClient) reconcileAutopilotNodePools(ctx context.Context, s *clusterState) error {
	if err := validateAutopilotPools(s.AutopilotPools); err != nil {
		return err
	}
	if c.autopilot == nil {
		if len(s.AutopilotPools) > 0 {
			return fmt.Errorf("Autopilot API client unavailable")
		}
		return nil
	}
	pools, err := c.autopilot.ListAutopilotPools(ctx, c.org, s.CloudspaceName)
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for _, p := range pools {
		existing[p.Name] = true
	}
	desired := map[string]bool{}
	for _, p := range s.AutopilotPools {
		desired[p.Name] = true
		if err := c.autopilot.WriteAutopilotPool(ctx, c.org, s.CloudspaceName, s.Region, p, existing[p.Name]); err != nil {
			return fmt.Errorf("reconcile Autopilot pool %s: %w", p.Name, err)
		}
	}
	for _, p := range pools {
		if !desired[p.Name] {
			if err := c.autopilot.DeleteAutopilotPool(ctx, c.org, p.Name); err != nil && !isNotFound(err) {
				return err
			}
		}
	}
	return nil
}
