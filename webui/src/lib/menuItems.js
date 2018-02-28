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
        name: 'Services',
        router: {
          name: 'ShowServices'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Instances',
        router: {
          name: 'ShowDependencies'
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
        name: 'Objects',
        router: {
          name: 'ShowObjects'
        }
      },
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
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'User Roles',
        router: {
          name: 'ShowUserRoles'
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
