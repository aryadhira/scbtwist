'use strict';

vm.currentMenu('Session');
vm.currentTitle("Session");
vm.breadcrumb([{ title: 'Operational', href: '#' },{ title: 'Session', href: viewModel.appName + 'page/session' }]);

viewModel.Session = new Object();
var ss = viewModel.Session;

ss.templateFilter = {
    search: ""
};

ss.TableColumns = ko.observableArray([
    { field: "status", title: "Status", headerAttributes:{ style:"text-align:center" }}, 
    { field: "loginid", title: "Username", headerAttributes:{ style:"text-align:center" }}, 
    { field: "created", title: "Created", headerAttributes:{ style:"text-align:center" }, template: '# if (created == "0001-01-01T00:00:00Z") {#-#} else {# #:moment(created).utc().format("DD-MMM-YYYY HH:mm:ss")# #}#' }, 
    { field: "expired", title: "Expired", headerAttributes:{ style:"text-align:center" }, template: '# if (expired == "0001-01-01T00:00:00Z") {#-#} else {# #:moment(expired).utc().format("DD-MMM-YYYY HH:mm:ss")# #}#' }, 
    { field: "duration", title: "Active In", headerAttributes:{ style:"text-align:center" }, attributes:{ class:"align-center" }, template: '#= kendo.toString(duration, "n2")# H' }, 
    { title: "Action", width: 80, headerAttributes:{ style:"text-align:center" }, attributes: { class: "align-center" }, template: "#if(status=='ACTIVE'){# <button data-value='#:_id #' onclick='ss.setexpired(\"#: _id #\", \"#: loginid #\")' name='expired' type='button' class='btn btn-sm btn-default btn-text-danger btn-stop tooltip tooltipster' title='Set Expired'><span class='fa fa-times'></span></button> #}else{# #}#" }]);

ss.contentIsLoading = ko.observable(false);
ss.selectedTableID = ko.observable("");
ss.filter = ko.mapping.fromJS(ss.templateFilter);

ss.refreshData = function () {
    ss.contentIsLoading(true);
    ss.generateGrid();
    $('.grid-session').data('kendoGrid').dataSource.read();
};

ss.setexpired = function (_id, username) {
    var param = { 
        _id: _id,
        username: username 
    };
    var activesession = '';

    toolkit.ajaxPost(viewModel.appName + "login/getsession", {}, function (ses) {
        if (!app.isFine(ses)) {
            return;
        }
        activesession = ses.data;
        toolkit.ajaxPost(viewModel.appName + "session/setexpired", param, function (res) {
            if (!app.isFine(res)) {
                return;
            }

            if(param._id == activesession) {
                toolkit.ajaxPost(viewModel.appName + "login/logout", {}, function (pes) {
                    if (!app.isFine(pes)) {
                        return;
                    }

                    location.href = viewModel.appName + 'page/login';
                });
            } else {
                ss.refreshData();
                location.reload();
            }
        });
    });
};

ss.generateGrid = function () {
    $(".grid-session").html("");
    $('.grid-session').kendoGrid({
        dataSource: {
            transport: {
                read: {
                    url: viewModel.appName + "session/getsession",
                    dataType: "json",
                    data: ko.mapping.toJS(ss.filter),
                    type: "POST",
                    contentType: "application/json; charset=utf-8",
                    success: function success(data) {
                        $(".grid-group>.k-grid-content-locked").css("height", $(".grid-group").data("kendoGrid").table.height());
                    }
                },
                parameterMap: function(options) {
                    return JSON.stringify(options);
                }
            },
            schema: {
                data: function data(res) {
                    ss.selectedTableID("show");
                    app.isFine(res);
                    return res.data.Data;
                },
                total: "data.total"
            },

            pageSize: 10,
            serverPaging: true, // enable server paging
            serverSorting: true
        },
        // selectable: "multiple, row",
        // change: ac.selectGridAccess,
        resizable: true,
        scrollable: true,
        // sortable: true,
        // filterable: true,
        pageable: {
            refresh: false,
            pageSizes: 10,
            buttonCount: 5
        },
        columns: ss.TableColumns()
    });
    ss.contentIsLoading(false);
};

$(function () {
    ss.generateGrid();
});