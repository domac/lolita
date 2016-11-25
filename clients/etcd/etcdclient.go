package etcd

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"time"
)

//Etcd客户端
type Client struct {
	client client.Client
}

var (
	cli Client
	ctx context.Context
)

const (
	//Watch事件名称
	Watch_Action_Create string = "create"
	Watch_Action_Set    string = "set"
	Watch_Action_Delete string = "delete"
)

func Init(endpoints []string) error {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second * 5,
	}

	var err error
	cli.client, err = client.New(cfg)
	if err != nil {
		log.Fatalf("connect to etcd err: %v", err)
		return err
	}
	log.Printf("etcd connect success : %v", endpoints)
	ctx = context.Background()
	return nil
}

func GetClient() *Client {
	return &cli
}

func GetContext() context.Context {
	return ctx
}

func (ec *Client) IsDirExist(dir string) bool {
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return false
	}

	return resp.Node.Dir
}

func (ec *Client) IsFileExist(file string) bool {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Get(ctx, file, nil)
	if err != nil {
		return false
	}
	return true
}

func (ec *Client) MakeDir(dir string) error {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Set(ctx, dir, "", &client.SetOptions{Dir: true})
	return err
}

func (ec *Client) CreateDir(dir string) error {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Create(ctx, dir, "")
	return err
}

func (ec *Client) Set(key, value string) error {
	kapi := client.NewKeysAPI(ec.client)
	_, err := kapi.Set(ctx, key, value, nil)
	return err
}

//从Etcd server获取值
func (ec *Client) Get(key string) (value string, err error) {
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, key, nil)
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

func (ec *Client) Delete(key string) (err error) {
	kapi := client.NewKeysAPI(ec.client)
	_, err = kapi.Delete(ctx, key, nil)
	if err != nil {
		return err
	}
	return nil
}

//列出目录的所有value
func (ec *Client) List(dir string) ([]string, error) {
	var values []string
	kapi := client.NewKeysAPI(ec.client)
	resp, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return values, err
	}

	for _, node := range resp.Node.Nodes {
		respNode, err := kapi.Get(ctx, node.Key, nil)
		if err != nil {
			return values, err
		}
		values = append(values, respNode.Node.Value)
	}
	return values, nil
}

//对指定的目录进行事件监听
func (ec *Client) CreateWatcher(dir string) (client.Watcher, error) {
	kapi := client.NewKeysAPI(ec.client)
	respGet, err := kapi.Get(ctx, dir, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("star watch %s after %d", dir, respGet.Index)
	w := kapi.Watcher(dir, &client.WatcherOptions{AfterIndex: respGet.Index,
		Recursive: true})
	return w, err
}
