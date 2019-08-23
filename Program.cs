using System;
using System.IO;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Linq;
using System.Runtime.InteropServices;
using PuppeteerSharp;

namespace Bl3AutoVip
{
    public static class Bl3AutoVip
    {   
        public static async Task WaitForXhr(this Page page)
        {
            await page.WaitForExpressionAsync(@"!window.XHR_COUNT");       
        }

        public static async Task GoToUrl(this Page page, string url)
        {
            Console.Write("Navigating to url '" + url + "' . . . . . ");
            try
            {
                await page.GoToAsync(url);
                await page.WaitForExpressionAsync("window.$");
                await page.EvaluateFunctionAsync
                (@"
                    function () {
                        if (typeof window.XHR_COUNT !== 'undefined')
                            return;
                        window.XHR_COUNT = 0;
                        var oldOpen = XMLHttpRequest.prototype.open;
                        XMLHttpRequest.prototype.open = function () {
                            window.XHR_COUNT += 1;
                            this.addEventListener('loadend', function (e) {
                                window.XHR_COUNT -= 1;
                            });
                            oldOpen.apply(this, arguments);
                        };
                    }
                ");
            }
            catch
            {
                Console.WriteLine("failed! Try again later.");
                throw;
            }
            Console.WriteLine("success!");
        }

        static async Task Main(string[] args)
        {
            Console.Write("Setting up . . . . . ");
            var driverFolder = RuntimeInformation.IsOSPlatform(OSPlatform.Windows)? "windows":
                               RuntimeInformation.IsOSPlatform(OSPlatform.Linux)? "linux":
                               "mac";
            var fetcherOptions = new BrowserFetcherOptions() { Path = Path.Join(".browsers" , driverFolder) };
            var fetcher = new BrowserFetcher(fetcherOptions);
            RevisionInfo execInfo;
            try
            {
                execInfo = await fetcher.DownloadAsync(BrowserFetcher.DefaultRevision);
                foreach (var oldVersion in fetcher.LocalRevisions().Where(x => x != BrowserFetcher.DefaultRevision))
                {
                    fetcher.Remove(oldVersion);
                }
            }
            catch
            {
                Console.WriteLine("failed! Try again later.");
                throw;
            }
            Console.WriteLine("success!");
            var browser = await Puppeteer.LaunchAsync(new LaunchOptions() { Headless = true, ExecutablePath = execInfo.ExecutablePath});
            try
            {
                var page = await browser.NewPageAsync();
                Console.Write("Enter username (email): ");
                var username = Console.ReadLine();
                Console.Write("Enter password        : ");
                var password = Console.ReadLine();
                await page.GoToUrl("https://borderlands.com/en-US/vip/");
                Console.Write("Logging in as '" + username + "' . . . . . ");
                try
                {
                    await page.WaitForExpressionAsync(@"window.$");
                    await page.WaitForExpressionAsync(@"$('[data-gtm-event=""vip-sign-up""]').length");
                    await page.EvaluateFunctionAsync
                    (@"
                        function () {
                            $('[data-gtm-event=""vip-sign-up""]').first().click();
                        }
                    ");
                    await page.WaitForExpressionAsync(@"$('[ng-click^=""Auth.showLoginModal""]').length");
                    await page.EvaluateFunctionAsync
                    (@"
                        function () {
                            $('[ng-click^=""Auth.showLoginModal""]').click();
                            $('[name=""username""]').val('" + username + @"').trigger('change');
                            $('[name=""password""]').val('" + password + @"').trigger('change');
                            $('[name=""form""]').submit();
                        }
                    ");
                    await page.WaitForXhr();
                    await page.WaitForNavigationAsync();
                }
                catch
                {
                    Console.WriteLine("failed! Try again later.");
                    throw;
                }
                Console.WriteLine("success!");
                await page.GoToUrl("https://borderlands.com/en-US/vip-codes/");
                await page.WaitForExpressionAsync(@"$('[widget-type=""activity-list""] iframe').length");
                await page.GoToUrl(await page.EvaluateExpressionAsync<string>(@"$('[widget-type=""activity-list""] iframe').prop('src')"));
                var widgetMap = await page.EvaluateFunctionAsync<Dictionary<string, string>>
                (@"
                    function () {
                        var x = {};
                        window.widgetConf.entries.map(function(e) {
                            x[e.activity.name] = e.link.widgetId;
                        });
                        return x;
                    }
                ");
                await page.GoToUrl("https://borderlands.com/en-US/profile/");
                await page.WaitForExpressionAsync(@"$('[widget-type=""activity-history""] iframe').length");
                await page.GoToUrl(await page.EvaluateExpressionAsync<string>(@"$('[widget-type=""activity-history""] iframe').prop('src')"));
                Console.Write("Getting redeemed codes . . . . . ");
                var redeemedCodes = new List<string>();
                try
                {
                    redeemedCodes = await page.EvaluateFunctionAsync<List<string>>
                    (@"
                        function () {
                            return JSON.parse($.ajax({
                                type: 'POST',
                                url: 'https://2kgames.crowdtwist.com/request?widgetId=9470',
                                contentType: 'application/javascript',
                                dataType: 'jsonp',
                                async: false,
                                data: JSON.stringify({
                                    model_data: {
                                        activity: {
                                            newest_activities: {
                                                properties: ['notes'],
                                                'query': {
                                                    'type': 'user_activities_me',
                                                    'args': {'row_start': 1,'row_end': 1000000
                                                    }
                                                }
                                            }
                                        }
                                    }
                                })
                            }).responseText).model_data.activity.newest_activities.map(function(y) {
                                return (y.notes || '').toUpperCase();
                            });
                        }
                    ");
                }
                catch
                {
                Console.WriteLine("failed! Try again later.");
                throw;
                }
                Console.WriteLine("success!");
                await page.GoToAsync("https://www.reddit.com/r/borderlands3/comments/bxgq5p/borderlands_vip_program_codes/");
                Console.Write("Getting new codes . . . . . ");
                var codesToRedeem = new Dictionary<string, List<string>>();
                try
                {
                    var allCodeRows = await page.EvaluateFunctionAsync<List<string[]>>
                    (@"
                        function () {
                            return Array.from(document.querySelectorAll('tbody tr')).map(function (x) { 
                                return Array.from(x.querySelectorAll('td')).map(function(y) {
                                    return y.textContent;
                                });
                            }).map(function (z) { 
                                return [z[3], /[Nn][Oo]/.test(z[2])? '': z[0]]; // code, type
                            });
                        }
                    ");
                    foreach (var codeRow in allCodeRows)
                    {
                        if (String.IsNullOrEmpty(codeRow[0]) || String.IsNullOrEmpty(codeRow[1]) || redeemedCodes.Contains(codeRow[1].ToUpper()))
                            continue;
                        if (!codesToRedeem.ContainsKey(codeRow[0].ToLower())) 
                        {
                            codesToRedeem[codeRow[0].ToLower()] = new List<string>();
                        }
                        codesToRedeem[codeRow[0].ToLower()].Add(codeRow[1].ToUpper());
                    }
                }
                catch
                {
                    Console.WriteLine("failed! Try again later.");
                    throw;
                }
                
                if (codesToRedeem.Count == 0)
                {
                    Console.WriteLine("No new codes at this time. Try again later.");
                    browser.Dispose();
                    return;
                }
                Console.WriteLine("success!");
                foreach(var keyValue in codesToRedeem)
                {
                    var codeTypeUrl = "https://2kgames.crowdtwist.com/widgets/t/code-redemption/" + widgetMap[widgetMap.Keys.First(x => x.Contains(keyValue.Key))] + "/";
                    await page.GoToUrl(codeTypeUrl);
                    await page.WaitForExpressionAsync("$('input').length");
                    foreach (var code in keyValue.Value)
                    {
                        Console.Write("Trying '" + keyValue.Key + "' code '" + code + "' . . . . . ");
                        try
                        {
                            await page.EvaluateFunctionAsync
                            (@"
                                function() {
                                    $('input').val('" + code + @"').trigger('change');
                                }
                            ");
                            await page.EvaluateFunctionAsync
                            (@"
                                function() {
                                    $('.submit-container button').click();
                                }
                            ");
                            await page.WaitForXhr();
                            var successOrFailureMessage = await page.EvaluateExpressionAsync<string>
                            (@"
                                $('[ng-show=""success""] .msg.ng-binding').text() + 
                                $('[ng-show=""modelError""] .msg.ng-binding').text()
                            ");
                            Console.WriteLine(successOrFailureMessage);
                        }
                        catch
                        {
                            Console.WriteLine("failed! Trying other codes.");
                        }
                    }
                }
            }
            catch
            {
                browser.Dispose();
                throw;
            }
            Console.WriteLine("Done with codes! Comeback when more codes are available.");
            browser.Dispose();
        }
    }
}
