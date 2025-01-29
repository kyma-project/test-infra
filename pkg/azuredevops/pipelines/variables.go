package pipelines

import "fmt"

func SetVariable(name string, value interface{}, isSecret bool, isOutput bool) {
	fmt.Printf("##vso[task.setvariable variable=%s;issecret=%v;isoutput=%v]%v\n", name, isSecret, isOutput, value)
}
