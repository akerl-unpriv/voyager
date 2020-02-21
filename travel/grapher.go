package travel

import (
	"github.com/akerl/voyager/v2/cartogram"

	"github.com/akerl/input/list"
)

// Grapher defines a graph resolution object for finding paths to accounts
type Grapher struct {
	Prompt list.Prompt
	Pack   cartogram.Pack
}

// ResolveAll selects all matching paths based on provided args
func (g *Grapher) ResolveAll(args, roleNames, profileNames []string) ([]Path, error) {
	logger.InfoMsgf("resolving all paths based on %v / %v / %v", args, roleNames, profileNames)

	tfs := cartogram.TagFilterSet{}
	if err := tfs.LoadFromArgs(args); err != nil {
		return []Path{}, err
	}
	accounts := g.Pack.Search(tfs)

	paths := make([]Path, len(accounts))

	for index, item := range accounts {
		path, err := g.filterPaths(item, roleNames, profileNames)
		if err != nil {
			return []Path{}, err
		}
		paths[index] = path
	}

	return paths, nil
}

// Resolve selects a valid path to the target account and role
func (g *Grapher) Resolve(args, roleNames, profileNames []string) (Path, error) {
	logger.InfoMsgf("resolving a path based on %v / %v / %v", args, roleNames, profileNames)
	account, err := g.selectTargetAccount(args)
	if err != nil {
		return Path{}, err
	}
	return g.filterPaths(account, roleNames, profileNames)
}

func (g *Grapher) filterPaths(account cartogram.Account, r, p []string) (Path, error) {
	paths, err := g.findAllPaths(account)
	if err != nil {
		return Path{}, err
	}

	paths, err = g.filterByRole(paths, r)
	if err != nil {
		return Path{}, err
	}

	paths, err = g.filterByProfile(paths, p)
	if err != nil {
		return Path{}, err
	}

	if len(paths) > 1 {
		logger.InfoMsg("multiple valid paths detected. Selecting the first option")
	}
	return paths[0], nil
}

func (g *Grapher) selectTargetAccount(args []string) (cartogram.Account, error) {
	logger.InfoMsgf("looking up account based on %v", args)
	return g.Pack.FindWithPrompt(args, g.Prompt)
}

func (g *Grapher) findAllPaths(account cartogram.Account) ([]Path, error) {
	var allPaths []Path

	for _, r := range account.Roles {
		paths, err := g.findPathToRole(account, r)
		if err != nil {
			return []Path{}, err
		}
		allPaths = append(allPaths, paths...)
	}

	logger.InfoMsgf("found %d paths", len(allPaths))
	return allPaths, nil
}

func (g *Grapher) findPathToRole(account cartogram.Account, role cartogram.Role) ([]Path, error) {
	var allPaths []Path

	for _, item := range role.Sources {
		account, role := item.Parse()
		if account != "" {
			newAccount, newRole, ok := g.pathIsViable(account, role)
			if !ok {
				continue
			}
			newPaths, err := g.findPathToRole(newAccount, newRole)
			if err != nil {
				return []Path{}, err
			}
			allPaths = append(allPaths, newPaths...)
		} else {
			allPaths = append(allPaths, Path{{
				Profile: item.Path,
			}})
		}
	}

	myHop := Hop{
		Role:    role.Name,
		Account: account.Account,
		Mfa:     role.Mfa,
		Region:  account.Region,
	}

	for i := range allPaths {
		allPaths[i] = append(allPaths[i], myHop)
	}
	return allPaths, nil
}

func (g *Grapher) pathIsViable(account, role string) (cartogram.Account, cartogram.Role, bool) {
	ok, accountObj := g.Pack.Lookup(account)
	if !ok {
		logger.DebugMsgf("found dead end due to missing account: %s", account)
		return cartogram.Account{}, cartogram.Role{}, false
	}
	ok, roleObj := accountObj.Roles.Lookup(role)
	if !ok {
		logger.DebugMsgf("found dead end due to missing role: %s/%s", account, role)
		return cartogram.Account{}, cartogram.Role{}, false
	}
	return accountObj, roleObj, true
}

func (g *Grapher) filterByRole(paths []Path, roleNames []string) ([]Path, error) {
	af := func(p Path) string { return p[len(p)-1].Role }

	allRoles := uniquePathAttributes(paths, af)

	role, err := list.WithInputSlice(
		g.Prompt,
		allRoles,
		roleNames,
		"Pick a role:",
	)
	if err != nil {
		return []Path{}, err
	}

	return filterPathsByAttribute(paths, role, af), nil
}

func (g *Grapher) filterByProfile(paths []Path, profileNames []string) ([]Path, error) {
	af := func(p Path) string { return p[0].Profile }

	allProfiles := uniquePathAttributes(paths, af)
	unionProfiles := sliceUnion(allProfiles, profileNames)

	profile, err := list.WithInputSlice(
		g.Prompt,
		allProfiles,
		unionProfiles,
		"Pick a profile:",
	)
	if err != nil {
		return []Path{}, err
	}

	return filterPathsByAttribute(paths, profile, af), nil
}
