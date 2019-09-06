function adminSubmit(event, action) {
    event.preventDefault();
    var pathname = window.location.pathname;
    const data = new URLSearchParams(new FormData(event.target));
    fetch(pathname, {
        method: 'POST',
        body: data,
        credentials: 'include'
    }).then(res => {
        if (res.status == 303) {
            return Promise.resolve();
        } else {
            return res.text()
                .then(message => {
                    return Promise.reject(message);
                });
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
        var action = event.target.getAttribute("data-action");
        var actionInfo = document.getElementById(action + "-info");
        actionInfo.classList.add('bg-danger');
        actionInfo.innerText = message;
    });
}

function initTemplateSupportCheck() {
    if (!('content' in document.createElement('template'))) {
        var templateSupportCheckElements = document.querySelectorAll('.template-support-check');
        for (i = 0; i < templateSupportCheckElements.length; i++) {
            templateSupportCheckElements[i].innerText =
                'HTML <template> tag not supported in browser, so editor cannot be shown.';
        }
    }
}

initTemplateSupportCheck();