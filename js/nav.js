var navTemplate = {
    init: function () {
        var mainNavbarNav = document.getElementById('main-navbar-nav');
        var currentNavAnchor = mainNavbarNav.querySelector('[href="' + window.location.pathname + '"]');
        if (currentNavAnchor != null) {
            currentNavAnchor.classList.add('active');
        }
    },

    toggleNavbar: function (event) {
        var mainNavbarNav = document.getElementById('main-navbar-nav');
        mainNavbarNav.classList.toggle('show');
    },

    toggleAdminMenu: function (event) {
        var adminDropdownMenu = document.getElementById('admin-dropdown-menu');
        adminDropdownMenu.classList.toggle('show');
    },
};

navTemplate.init();