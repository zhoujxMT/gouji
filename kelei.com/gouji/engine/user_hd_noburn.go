/*
玩家-操作-不烧牌
*/

package engine

//不烧牌
func NoBurn(args []string) *string {
	res := ForceCheckCard(args)
	return res
}
