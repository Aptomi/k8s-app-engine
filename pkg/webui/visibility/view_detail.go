package visibility

type item struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

// TODO: UI may be broken now because lint forced changing UserId -> UserID
type detail struct {
	UserID          string
	Users           []*item
	Services        []*item
	Dependencies    []*item
	AllDependencies []*item
	Views           []*item
	//	Summary         diff.RevisionSummary
}

// NewDetails returns detail object
func NewDetails(userID string) interface{} {
	/*
		revision := resolve.LoadRevision()
		summary := diff.GetSummary(revision)
	*/
	r := detail{
		userID,
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		//		summary,
	}

	/*
		// Users
		userIds := make([]string, 0)
		for userID := range revision.UserLoader.LoadUsersAll().Users {
			userIds = append(userIds, userID)
		}

		sort.Strings(userIds)

		if len(userIds) > 1 {
			r.Users = append([]*item{{"all", "All"}}, r.Users...)
		}
		for _, userID := range userIds {
			r.Users = append(r.Users, &item{userID, revision.UserLoader.LoadUserByName(userID).Name})
		}

		// Dependencies
		depIds := make([]string, 0)
		deps := revision.Policy.Dependencies.DependenciesByID
		for depID, dep := range deps {
			if dep.UserID != userID {
				continue
			}

			depIds = append(depIds, depID)
		}

		sort.Strings(depIds)

		if len(depIds) > 1 {
			r.Dependencies = append([]*item{{"all", "All"}}, r.Dependencies...)
		}
		for _, depID := range depIds {
			r.Dependencies = append(r.Dependencies, &item{depID, deps[depID].GetID()})
		}

		allDepIds := make([]string, 0)
		for depID := range deps {
			allDepIds = append(allDepIds, depID)
		}

		sort.Strings(allDepIds)

		if len(allDepIds) > 1 {
			r.AllDependencies = append([]*item{{"all", "All"}}, r.AllDependencies...)
		}
		for _, depID := range allDepIds {
			r.AllDependencies = append(r.AllDependencies, &item{depID, deps[depID].GetID()})
		}

		// Services
		svcIds := make([]string, 0)
		for svcID, svc := range revision.Policy.Services {
			if svc.Owner != userID {
				continue
			}
			svcIds = append(svcIds, svcID)
		}

		sort.Strings(svcIds)

		for _, svcID := range svcIds {
			r.Services = append(r.Services, &item{svcID, revision.Policy.Services[svcID].Name})
		}

		if len(r.Dependencies) > 0 {
			r.Views = append(r.Views, &item{"consumer", "Service Consumer"})
		}
		if len(r.Services) > 0 {
			r.Views = append(r.Views, &item{"service", "Service Owner"})
		}

		// TODO: this will have to be changed when we implement roles & ACLs
		if revision.UserLoader.LoadUserByName(userID).Labels["global_ops"] == "true" {
			r.Views = append(r.Views, &item{"globalops", "Global IT/Ops"})
		}
	*/
	return r
}
