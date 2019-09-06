function twoDigits(num) {
  return num < 10 ? "0" + num : num;
}

function formatDate(utcDate) {
  var localDate = new Date(utcDate);
  var year = localDate.getFullYear();
  var month = localDate.getMonth() + 1;
  var date = localDate.getDate();
  var hours = localDate.getHours();
  var minutes = localDate.getMinutes();
  var seconds = localDate.getSeconds();
  return twoDigits(year) + "/" +
    twoDigits(month) + "/" +
    twoDigits(date) + " " +
    twoDigits(hours) + ":" +
    twoDigits(minutes) + ":" +
    twoDigits(seconds);
}

function initCurrentNavbarPage() {
  var mainNavbarNav = document.getElementById('main-navbar-nav');
  var currentNavAnchor = mainNavbarNav.querySelector('[href="' + window.location.pathname + '"]');
  if (currentNavAnchor != null) {
    currentNavAnchor.classList.add('active');
  }
}

function toggleNavbar(event) {
  var mainNavbarNav = document.getElementById('main-navbar-nav');
  mainNavbarNav.classList.toggle('show');
}

function toggleAdminMenu(event) {
  var adminDropdownMenu = document.getElementById('admin-dropdown-menu');
  adminDropdownMenu.classList.toggle('show');
}

function initFirstTab() {
  if (document.getElementById('main-tabs-tabs') == null) {
    return;
  }
  var tabs = document.getElementById('main-tabs-tabs').querySelectorAll(".nav-link");
  for (var i = 0; i < tabs.length; i++) {
    if (tabs[i].hash === window.location.hash) {
      activateTab(tabs[i]);
      return;
    }
  }
  if (tabs.length > 0) {
    activateTab(tabs[0]);
  }
}

function tabClick(e) {
  e.preventDefault();
  window.location.hash = e.srcElement.hash;
  activateTab(e.srcElement);
}

function activateTab(clickedTab) {
  if (clickedTab.classList.contains('active')) {
    return;
  }

  var tabs = document.getElementById('main-tabs-tabs').querySelectorAll(".nav-link");
  for (var i = 0; i < tabs.length; i++) {
    var selectedTab = tabs[i].id === clickedTab.id;
    tabs[i].classList.toggle('active', selectedTab);
  }

  var tabContents = document.getElementById('main-tabs-content').querySelectorAll('.tab-pane');
  for (var i = 0; i < tabContents.length; i++) {
    var selectedTabContent = "#" + tabContents[i].id === clickedTab.hash;
    tabContents[i].classList.toggle('active', selectedTabContent);
    tabContents[i].classList.toggle('show', selectedTabContent);
  }
}

function initBottomMessages() {
  var timesMessageElement = document.getElementById('times-message');
  if (timesMessageElement != null) {
    var messages = timesMessageElement.querySelector(".messages").children;
    var times = timesMessageElement.querySelector(".times").children;
    var formattedTimesMessage = "";
    for (var i = 0; i < messages.length; i++) {
      formattedTimesMessage += messages[i].innerText;
      if (times && i < times.length) {
        formattedTimesMessage += formatDate(times[i].innerText);
      }
    }
    document.getElementById('times-message').innerText = formattedTimesMessage;
  }

  var pageLoadMessageElement = document.getElementById('page-load-message');
  var pageLoadTime = pageLoadMessageElement.innerText;
  pageLoadMessageElement.innerText = 'Page loaded at ' + formatDate(pageLoadTime);
}

initCurrentNavbarPage();
initFirstTab();
initBottomMessages();