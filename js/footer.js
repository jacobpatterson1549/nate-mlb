var footerTemplate = {
    twoDigits: function (num) {
        return num < 10 ? "0" + num : num;
    },

    formatDate: function (utcDate) {
        var localDate = new Date(utcDate);
        var year = localDate.getFullYear();
        var month = localDate.getMonth() + 1;
        var date = localDate.getDate();
        var hours = localDate.getHours();
        var minutes = localDate.getMinutes();
        var seconds = localDate.getSeconds();
        return footerTemplate.twoDigits(year) + "/" +
            footerTemplate.twoDigits(month) + "/" +
            footerTemplate.twoDigits(date) + " " +
            footerTemplate.twoDigits(hours) + ":" +
            footerTemplate.twoDigits(minutes) + ":" +
            footerTemplate.twoDigits(seconds);
    },

    initTimesMessage: function () {
        var timesMessageElement = document.getElementById('times-message');
        if (timesMessageElement == null) {
            return;
        }
        var messages = timesMessageElement.querySelector(".messages").children;
        var times = timesMessageElement.querySelector(".times").children;
        var formattedTimesMessage = "";
        for (var i = 0; i < messages.length; i++) {
            if (i > 0) {
                formattedTimesMessage += " ";
            }
            formattedTimesMessage += messages[i].innerText;
            if (times && i < times.length) {
                formattedTimesMessage += " " + footerTemplate.formatDate(times[i].innerText);
            }
        }
        timesMessageElement.innerText = formattedTimesMessage;
    },

    initPageLoadMessage: function () {
        var pageLoadMessageElement = document.getElementById('page-load-message');
        var pageLoadTime = pageLoadMessageElement.innerText;
        pageLoadMessageElement.innerText = 'Page loaded at ' + footerTemplate.formatDate(pageLoadTime);
    },

    init: function () {
        footerTemplate.initTimesMessage();
        footerTemplate.initPageLoadMessage();
    },
};

footerTemplate.init();