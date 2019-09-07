var tabsTemplate = {
    init: function () {
      if (document.getElementById('main-tabs-tabs') == null) {
        return;
      }
      var tabs = document.getElementById('main-tabs-tabs').querySelectorAll(".nav-link");
      for (var i = 0; i < tabs.length; i++) {
        if (tabs[i].hash === window.location.hash) {
          tabsTemplate.activateTab(tabs[i]);
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
      var tabs = document.getElementById('main-tabs-tabs').querySelectorAll(".nav-link");
      for (var i = 0; i < tabs.length; i++) {
        var selectedTab = tabs[i].id === clickedTab.id;
        tabs[i].classList.toggle('active', selectedTab);
      }
      var tabContents = document.getElementById('main-tabs-content').querySelectorAll('.tab-pane');
      for (i = 0; i < tabContents.length; i++) {
        var selectedTabContent = "#" + tabContents[i].id === clickedTab.hash;
        tabContents[i].classList.toggle('active', selectedTabContent);
        tabContents[i].classList.toggle('show', selectedTabContent);
      }
    },
  };
  
  tabsTemplate.init();