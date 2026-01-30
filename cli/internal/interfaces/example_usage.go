package interfaces

// This file contains example usage patterns for the anime CLI interfaces.
// These are not meant to be run, but to demonstrate best practices.

/*
// Example 1: Simple SSH operation with dependency injection

func deployApplication(client SSHClient, appPath string) error {
    // Check if deployment directory exists
    output, err := client.RunCommand("test -d " + appPath + " && echo exists")
    if err != nil {
        return err
    }

    if output != "exists\n" {
        // Create directory
        _, err = client.RunCommand("mkdir -p " + appPath)
        if err != nil {
            return err
        }
    }

    // Upload deployment script
    script := "#!/bin/bash\necho 'Deploying application...'\n"
    if err := client.UploadString(script, appPath+"/deploy.sh"); err != nil {
        return err
    }

    // Make it executable
    if err := client.MakeExecutable(appPath + "/deploy.sh"); err != nil {
        return err
    }

    // Execute deployment
    _, err = client.RunCommand(appPath + "/deploy.sh")
    return err
}

// Example 2: Installing modules with progress monitoring

func installWithProgress(installer Installer, modules []string) error {
    // Configure parallel installation
    installer.SetParallel(true)
    installer.SetJobs(4)

    // Monitor progress in a goroutine
    done := make(chan bool)
    go func() {
        for update := range installer.GetProgressChannel() {
            if update.Error != nil {
                log.Printf("Error in %s: %v", update.Module, update.Error)
            } else {
                log.Printf("[%s] %s", update.Module, update.Status)
                if update.Output != "" {
                    log.Printf("  %s", update.Output)
                }
            }
        }
        done <- true
    }()

    // Start installation
    err := installer.Install(modules)

    // Wait for progress monitoring to complete
    <-done

    return err
}

// Example 3: Complete deployment service using multiple interfaces

type DeploymentOrchestrator struct {
    ssh     SSHClient
    install Installer
    source  SourceController
    pkg     PackageManager
}

func NewDeploymentOrchestrator(
    ssh SSHClient,
    install Installer,
    source SourceController,
    pkg PackageManager,
) *DeploymentOrchestrator {
    return &DeploymentOrchestrator{
        ssh:     ssh,
        install: install,
        source:  source,
        pkg:     pkg,
    }
}

func (d *DeploymentOrchestrator) FullDeployment(modules []string, packages []string) error {
    // Step 1: Test SSH connection
    if err := d.install.TestConnection(); err != nil {
        return fmt.Errorf("connection test failed: %w", err)
    }

    // Step 2: Link source repository
    if err := d.source.Link("production/app"); err != nil {
        return fmt.Errorf("failed to link source: %w", err)
    }

    // Step 3: Push source code
    if err := d.source.Push(); err != nil {
        return fmt.Errorf("failed to push source: %w", err)
    }

    // Step 4: Install system modules
    if err := d.install.Install(modules); err != nil {
        return fmt.Errorf("module installation failed: %w", err)
    }

    // Step 5: Install packages
    for _, pkg := range packages {
        if err := d.pkg.Install(pkg, false, false); err != nil {
            return fmt.Errorf("package installation failed: %w", err)
        }
    }

    return nil
}

// Example 4: Testing with mocks

type MockSSHClient struct {
    commands []string
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
    m.commands = append(m.commands, cmd)
    return "success", nil
}

func (m *MockSSHClient) RunCommandWithProgress(cmd string, progress chan<- string) error {
    m.commands = append(m.commands, cmd)
    progress <- "output line 1"
    progress <- "output line 2"
    return nil
}

func (m *MockSSHClient) UploadString(content, path string) error {
    m.commands = append(m.commands, "upload:"+path)
    return nil
}

func (m *MockSSHClient) MakeExecutable(path string) error {
    m.commands = append(m.commands, "chmod:"+path)
    return nil
}

func (m *MockSSHClient) Close() error {
    return nil
}

func TestDeployment() {
    mock := &MockSSHClient{}

    err := deployApplication(mock, "/opt/app")
    if err != nil {
        panic(err)
    }

    // Verify commands were executed
    fmt.Printf("Executed %d commands\n", len(mock.commands))
    for _, cmd := range mock.commands {
        fmt.Printf("  - %s\n", cmd)
    }
}

// Example 5: Factory pattern for creating interface implementations

type ClientFactory struct {
    host    string
    user    string
    keyPath string
}

func NewClientFactory(host, user, keyPath string) *ClientFactory {
    return &ClientFactory{
        host:    host,
        user:    user,
        keyPath: keyPath,
    }
}

func (f *ClientFactory) CreateSSHClient() (SSHClient, error) {
    // In a real implementation, this would create ssh.Client
    // return ssh.NewClient(f.host, f.user, f.keyPath)
    return nil, nil
}

func (f *ClientFactory) CreateInstaller(client SSHClient) Installer {
    // In a real implementation, this would create installer.Installer
    // return installer.New(client.(*ssh.Client))
    return nil
}

func (f *ClientFactory) CreateSourceController(config *source.Config) SourceController {
    // In a real implementation, this would create source.Controller
    // return source.NewController(f.host, config)
    return nil
}

func (f *ClientFactory) CreatePackageManager(config *pkg.Config) PackageManager {
    // In a real implementation, this would create pkg.Manager
    // return pkg.NewManager(f.host, config)
    return nil
}

// Example 6: Chain of responsibility with interfaces

type DeploymentStep interface {
    Execute(ctx *DeploymentContext) error
    Name() string
}

type DeploymentContext struct {
    SSH       SSHClient
    Installer Installer
    Source    SourceController
    Packages  PackageManager
    Data      map[string]interface{}
}

type ConnectivityCheckStep struct{}

func (s *ConnectivityCheckStep) Execute(ctx *DeploymentContext) error {
    return ctx.Installer.TestConnection()
}

func (s *ConnectivityCheckStep) Name() string {
    return "Connectivity Check"
}

type SourceSyncStep struct{}

func (s *SourceSyncStep) Execute(ctx *DeploymentContext) error {
    if err := ctx.Source.Status(); err != nil {
        return err
    }
    return ctx.Source.Push()
}

func (s *SourceSyncStep) Name() string {
    return "Source Sync"
}

type ModuleInstallStep struct {
    modules []string
}

func (s *ModuleInstallStep) Execute(ctx *DeploymentContext) error {
    return ctx.Installer.Install(s.modules)
}

func (s *ModuleInstallStep) Name() string {
    return "Module Installation"
}

type DeploymentPipeline struct {
    steps []DeploymentStep
    ctx   *DeploymentContext
}

func NewDeploymentPipeline(ctx *DeploymentContext) *DeploymentPipeline {
    return &DeploymentPipeline{
        steps: []DeploymentStep{},
        ctx:   ctx,
    }
}

func (p *DeploymentPipeline) AddStep(step DeploymentStep) {
    p.steps = append(p.steps, step)
}

func (p *DeploymentPipeline) Execute() error {
    for _, step := range p.steps {
        log.Printf("Executing step: %s", step.Name())
        if err := step.Execute(p.ctx); err != nil {
            return fmt.Errorf("step %s failed: %w", step.Name(), err)
        }
    }
    return nil
}

// Usage:
func RunDeploymentPipeline(ctx *DeploymentContext, modules []string) error {
    pipeline := NewDeploymentPipeline(ctx)
    pipeline.AddStep(&ConnectivityCheckStep{})
    pipeline.AddStep(&SourceSyncStep{})
    pipeline.AddStep(&ModuleInstallStep{modules: modules})
    return pipeline.Execute()
}

*/
