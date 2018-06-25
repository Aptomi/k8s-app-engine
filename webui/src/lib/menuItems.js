module.exports = [
  {
    type: 'item',
    isHeader: true,
    name: 'MAIN NAVIGATION'
  },
  {
    type: 'tree',
    icon: 'fa fa-dashboard',
    name: 'Objects',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Bundles',
        router: {
          name: 'ShowBundles'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Services',
        router: {
          name: 'ShowServices'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Claims',
        router: {
          name: 'ShowClaims'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Rules',
        router: {
          name: 'ShowRules'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Clusters',
        router: {
          name: 'ShowClusters'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Users',
        router: {
          name: 'ShowUserRoles'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'All Objects',
        router: {
          name: 'ShowCatalog'
        }
      }
    ]
  },
  {
    type: 'tree',
    icon: 'fa fa-dashboard',
    name: 'Deployment',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Browser',
        router: {
          name: 'BrowsePolicy'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Audit Log',
        router: {
          name: 'ShowAuditLog'
        }
      }
    ]
  },
  {
    type: 'Help',
    icon: 'fa fa-book',
    name: 'Help',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Web Site',
        router: {
          name: 'Website'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Documentation',
        router: {
          name: 'Documentation'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Slack',
        router: {
          name: 'Slack'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Github',
        router: {
          name: 'Github'
        }
      }
    ]
  }

]
