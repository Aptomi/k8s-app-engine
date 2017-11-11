module.exports = [
  {
    type: 'item',
    isHeader: true,
    name: 'MAIN NAVIGATION'
  },
  {
    type: 'tree',
    icon: 'fa fa-dashboard',
    name: 'Policy',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Objects',
        router: {
          name: 'Hello'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Browser',
        router: {
          name: 'Sample'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Audit Log',
        router: {
          name: 'DashboardV1'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Instances',
        router: {
          name: 'DashboardV2'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'User Roles',
        router: {
          name: 'InfoBoxExample'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Chart.js',
        router: {
          name: 'ChartExample'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Alert',
        router: {
          name: 'AlertExample'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Modal',
        router: {
          name: 'ModalExample'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'APIExample',
        router: {
          name: 'APIExample'
        }
      }
    ]
  },
  {
    type: 'item',
    icon: 'fa fa-th',
    name: 'Widgets',
    badge: {
      type: 'String',
      data: 'new'
    },
    router: {
      name: 'WidgetsExample'
    }
  },
  {
    type: 'tree',
    icon: 'fa fa-laptop',
    name: 'UI Elements',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'General',
        router: {
          name: 'General'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Icons',
        router: {
          name: 'Icons'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Buttons',
        router: {
          name: 'Buttons'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Sliders',
        router: {
          name: 'Sliders'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Timeline',
        router: {
          name: 'Timeline'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Modals',
        router: {
          name: 'Modals'
        }
      }
    ]
  },
  {
    type: 'tree',
    icon: 'fa fa-edit',
    name: 'Forms',
    items: [
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'General Elements',
        router: {
          name: 'GeneralElements'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Advanced Elements',
        router: {
          name: 'AdvancedElements'
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
        name: 'Documentation',
        router: {
          name: 'AdvancedElements'
        }
      },
      {
        type: 'item',
        icon: 'fa fa-circle-o',
        name: 'Slack',
        router: {
          name: 'AdvancedElements'
        }
      }
    ]
  }

]
