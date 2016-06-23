package gmvc

import (
    "strings"
    "fmt"
    "sync"
)

/*
Tree is a structred data storage, best for configuration
*/
type Tree struct {
    name string
    branches map[string]*Tree
    value interface{}
    locker *sync.RWMutex
    parent *Tree
}

func NewTree() *Tree {
    return &Tree{name: "", locker: &sync.RWMutex{}}
}

/*
LoadYaml reads data from a yaml file,
repl means whether to replace or keep the old value
*/
func (t *Tree) LoadYaml(key, file string, repl bool) error {
    var yml interface{}
    if err := LoadYaml(&yml, file); err != nil {
        return err
    }
    tree := t.prepare(key)
    tree.loadValue(yml, repl)
    return nil
}

/*
LoadJson read data from a json file,
repl means whether to replace or keep the old value
*/
func (t *Tree) LoadJson(key, file string, repl bool) error {
    var json interface{}
    if err := LoadJson(&json, file); err != nil {
        return err
    }
    tree := t.prepare(key)
    tree.loadValue(json, repl)
    return nil
}

func (t *Tree) LoadTree(key string, data *Tree) {
    tree := t.prepare(key)
    tree.branches = data.branches
    if tree.branches == nil {
        tree.branches = data.branches
    } else {
        for k, v := range data.branches {
            tree.branches[k] = v
        }
    }
}


func (t *Tree) loadValue(val interface{}, repl bool) {
    if v, ok := val.(map[interface{}]interface{}); ok {
        t.loadBranches(v, nil, repl)
    } else if v, ok := val.([]interface{}); ok {
        t.loadBranches(nil, v, repl)
    } else if repl || t.value == nil {
        t.value = val
    }
}

func (t *Tree) loadBranches(m map[interface{}]interface{}, arr []interface{}, repl bool) {
    if t.branches == nil {
        t.branches = make(map[string]*Tree)
    }
    for k, v := range m {
        t.loadBranch(fmt.Sprintf("%v", k), v, repl)
    }
    for k, v := range arr {
        t.loadBranch(fmt.Sprintf("%d", k), v, repl)
    }
}

func (t *Tree) loadBranch(key string, val interface{}, repl bool) {
    tree, has := t.branches[key]
    if !has {
        tree = &Tree{name: key, locker: t.locker, parent: t}
        t.branches[key] = tree
    }
    tree.loadValue(val, repl)
}

func (t *Tree) find(key string) *Tree {
    if key == "" {
        return t
    }
    t.locker.RLock()
    defer t.locker.RUnlock()
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        var has bool
        if current.branches == nil {
            return nil
        }
        if current, has = current.branches[name]; !has {
            return nil
        }
    }
    return current
}

/*
Get returns value find by key, key is a path divided by "." dot notation
*/
func (t *Tree) Get(key string) interface{} {
    tree := t.find(key)
    if tree == nil {
        return nil
    }
    return tree.value
}

func (t *Tree) prepare(key string) *Tree {
    if key == "" {
        return t
    }
    t.locker.Lock()
    defer t.locker.Unlock()
    current := t
    nodes := strings.Split(
        strings.ToLower(strings.Trim(key, ".")), ".")
    for _, name := range nodes {
        var tree *Tree
        var has bool
        if current.branches == nil {
            current.branches = make(map[string]*Tree)
            has = false
        } else {
            tree, has = current.branches[name]
        }
        if !has {
            tree = &Tree{name: name, locker: t.locker, parent: t}
            current.branches[name] = tree
        }
        current = tree
    }
    return current
}

func (t *Tree) Set(key string, val interface{}) {
    tree := t.prepare(key)
    tree.value = val
}

func (t *Tree) Add(key string, val interface{}) bool {
    if tree := t.prepare(key); tree.value == nil {
        tree.value = val
        return true
    }
    return false
}

func (t *Tree) Tree(key string) *Tree {
    tree := t.find(key)
    if tree == nil {
        return nil
    }
    return tree
}

func (t *Tree) Branches() map[string]*Tree {
    return t.branches
}

func (t *Tree) NodeNum(key string) int {
    tree := t.find(key)
    return len(tree.branches)
}

func (t *Tree) Clear() {
    t.branches = nil
    t.value = nil
}

func (t *Tree) Int(key string) (int, bool) {
    if v := t.Get(key); v != nil {
        i, ok := v.(int)
        return i, ok
    }
    return 0, false
}

func (t *Tree) Int64(key string) (int64, bool) {
    if v := t.Get(key); v != nil {
        if i, ok := v.(int64); ok {
            return i, true
        }
        if i, ok := v.(int); ok {
            return int64(i), ok
        }
    }
    return 0, false
}

func (t *Tree) Float64(key string) (float64, bool) {
    if v := t.Get(key); v != nil {
        f, ok := v.(float64)
        return f, ok
    }
    return 0, false
}

func (t *Tree) Bool(key string) (bool, bool) {
    if v := t.Get(key); v != nil {
        s, ok := v.(bool)
        return s, ok
    }
    return false, false
}

func (t *Tree) String(key string) (string, bool) {
    if v := t.Get(key); v != nil {
        s, ok := v.(string)
        return s, ok
    }
    return "", false
}
