var adminTab = {
    submit: function (event) {
        event.preventDefault();
        var pathname = window.location.pathname;
        var data = new URLSearchParams(new FormData(event.target));
        fetch(pathname, {
            method: 'POST',
            body: data,
            credentials: 'include'
        }).then(async res => {
            if (res.status == 303) {
                return Promise.resolve();
            } else {
                var message = await res.text();
                return Promise.reject(message);
            }
        }).then(() => {
            if (window.PasswordCredential) {
                var c = new PasswordCredential(event.target);
                return navigator.credentials.store(c);
            } else {
                return Promise.resolve();
            }
        }).then(() => {
            location.reload();
        }).catch(message => {
            var action = event.target.getAttribute('data-action');
            var actionInfo = document.getElementById(action + '-info');
            actionInfo.classList.add('bg-danger');
            actionInfo.innerText = message;
        });
    },

    init: function () {
        if (!('content' in document.createElement('template'))) {
            var templateSupportCheckElements = document.querySelectorAll('.template-support-check');
            for (var i = 0; i < templateSupportCheckElements.length; i++) {
                templateSupportCheckElements[i].innerText =
                    'HTML <template> tag not supported in browser, so editor cannot be shown.';
            }
        }
    },
};

adminTab.init();