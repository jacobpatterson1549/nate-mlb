var tabsTemplate = {
    init: function () {
      if (document.getElementById('main-tabs-tabs') == null) {
        return;
      }
      var tabs = document.getElementById('main-tabs-tabs').querySelectorAll('.nav-link');
      for (var tab of tabs) {
        if (tab.hash === window.location.hash) {
          tabsTemplate.activateTab(tab);
          return;
        }
      }
      if (tabs.length > 0) {
        tabsTemplate.activateTab(tabs[0]);
      }
    },
  
    tabClick: function (event) {
      event.preventDefault();
      window.location.hash = event.srcElement.hash;
      tabsTemplate.activateTab(event.srcElement);
    },
  
    activateTab: function (clickedTab) {
      if (clickedTab.classList.contains('active')) {
        return;
      }
      var tabs = document.getElementById('main-tabs-tabs').querySelectorAll('.nav-link');
      for (var tab of tabs) {
        var selectedTab = tab.id === clickedTab.id;
        tab.classList.toggle('active', selectedTab);
      }
      var tabContents = document.getElementById('main-tabs-content').querySelectorAll('.tab-pane');
      for (var tabContent of tabContents) {
        var selectedTabContent = '#' + tabContent.id === clickedTab.hash;
        tabContent.classList.toggle('active', selectedTabContent);
        tabContent.classList.toggle('show', selectedTabContent);
      }
    },
  };
  
  tabsTemplate.init();