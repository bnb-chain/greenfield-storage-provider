package main

import (
	"os"
	"text/template"
)

const composeTemplate = `
services:
  mysql:
    image: {{.MySQLImage}}
    networks:
      - mechain-network
    container_name: sp-mysql
    volumes:
      - db-data:/var/lib/mysql
    environment:
      MYSQL_ROOT_PASSWORD: mechain
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
  init:
    container_name: init-sp
    image: "{{$.Image}}"
    networks:
      - mechain-network    
    depends_on:
      mysql:
        condition: service_healthy
    volumes:
      - "{{$.ProjectBasePath}}/deployment/dockerup/sp.json:/workspace/sp.json:Z"
      - "{{$.ProjectBasePath}}/deployment/dockerup:/workspace/deployment/dockerup:Z"
      - "local-env:/workspace/deployment/dockerup/local_env"
    working_dir: "/workspace/deployment/dockerup"
    command: >
      bash -c "
      rm -f init_done &&
      mkdir -p /workspace/build &&
      cp /usr/bin/mechain-sp /workspace/build/mechain-sp &&
      bash localup.sh --generate /workspace/sp.json root mechain mysql:3306 && 
      bash localup.sh --reset &&
      touch init_done && 
      sleep infinity
      "
    healthcheck:
      test: ["CMD-SHELL", "test -f /workspace/deployment/dockerup/init_done && echo 'OK' || exit 1"]
      interval: 10s
      retries: 5
    restart: "on-failure"
{{- range .Nodes }}
  spnode-{{.NodeIndex}}:
    container_name: mechain-sp-{{.NodeIndex}}
    depends_on:
      init:
        condition: service_healthy
    image: "{{$.Image}}"
    networks:
      - mechain-network
    ports:
      - "{{.GRPCPort}}:{{$.BasePorts.GRPCPort}}"
      - "{{.P2PPort}}:{{$.BasePorts.P2PPort}}"
      - "{{.MetricPort}}:{{$.BasePorts.MetricPort}}"
      - "{{.PprofPort}}:{{$.BasePorts.PprofPort}}"
      - "{{.ProbePort}}:{{$.BasePorts.ProbePort}}"
    volumes:
      - "local-env:/app"
    working_dir: "/app/sp{{.NodeIndex}}/"
    command: >
      ./mechain-sp{{.NodeIndex}} --config config.toml </dev/null >log.txt 2>&1 &
{{- end }}
volumes:
  db-data:
  local-env:
networks:
  mechain-network:
    external: true
`

type basePorts struct {
	GRPCPort   int
	P2PPort    int
	MetricPort int
	PprofPort  int
	ProbePort  int
}
type NodeConfig struct {
	NodeIndex int
	basePorts
}

type ComposeConfig struct {
	Nodes           []NodeConfig
	Image           string
	MySQLImage      string
	ProjectBasePath string
	BasePorts       basePorts
}

func main() {
	numNodes := 8

	bp := basePorts{
		GRPCPort: 9033,
		P2PPort:  9063,
	}
	bp.MetricPort = bp.GRPCPort + 367
	bp.PprofPort = bp.GRPCPort + 368
	bp.ProbePort = bp.GRPCPort + 369

	nodes := make([]NodeConfig, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = NodeConfig{
			NodeIndex: i,
			basePorts: basePorts{
				GRPCPort:   i + bp.GRPCPort,
				P2PPort:    i + bp.P2PPort,
				MetricPort: i*1000 + bp.MetricPort,
				PprofPort:  i*1000 + bp.PprofPort,
				ProbePort:  i*1000 + bp.ProbePort,
			},
		}
	}
	composeConfig := ComposeConfig{
		Nodes:           nodes,
		Image:           "zkmelabs/mechain-storage-provider",
		MySQLImage:      "mysql:8",
		ProjectBasePath: ".",
		BasePorts:       bp,
	}

	tpl, err := template.New("docker-compose").Parse(composeTemplate)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("docker-compose.yml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = tpl.Execute(file, composeConfig)
	if err != nil {
		panic(err)
	}

	println("Docker Compose file generated successfully!")
}
