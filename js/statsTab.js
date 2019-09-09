var statsTab = {
    init: function () {
        var statsAdminLinks = document.querySelectorAll(".stats-admin-link");
        for (var i = 0; i < statsAdminLinks.length; i++) {
            statsAdminLinks[i].href = location.pathname + statsAdminLinks[i].pathname;
        }
    },
};

statsTab.init();