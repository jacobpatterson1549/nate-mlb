var statsTab = {
    init: function () {
        var statsAdminLinks = document.querySelectorAll('.stats-admin-link');
        for (var statsAdminLink of statsAdminLinks) {
            statsAdminLink.href = location.pathname + statsAdminLink.getAttribute('data-relative-path');
        }
    },
};

statsTab.init();