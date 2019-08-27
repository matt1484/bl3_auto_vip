using System;
using System.Threading;
using System.Runtime.InteropServices;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Linq;
using System.Net.Http;
using System.Net;
using AngleSharp;
using AngleSharp.Dom;
using Newtonsoft.Json;
using CommandLine;

namespace Bl3AutoVip
{
    public class Bl3AutoVipArgs
    {
        [Option('e', "email", Required = false, HelpText = "Email for shift account.")]
        public string Email { get; set; } = "";
        [Option('p', "password", Required = false, HelpText = "Password for shift account.")]
        public string Password { get; set; } = "";
    }

    public static class Bl3AutoVip
    {   
        public static void Exit(Exception ex = null)
        {
            if (ex != null)
            {
                Console.WriteLine("failed! Had error: ");
                Console.WriteLine(ex.ToString());
            }
            Console.Write("Exiting in ");
            for (int i = 5; i > 0; i--)
            {
                Console.Write(i + " ");
                Thread.Sleep(1000);
            }
            Console.WriteLine("");
        }
        static async Task Main(string[] args)
        {
            try
            {
                Bl3AutoVipArgs cmdArgs = new Bl3AutoVipArgs();
                var argsResult = Parser.Default.ParseArguments<Bl3AutoVipArgs>(args).WithParsed<Bl3AutoVipArgs>(x => { cmdArgs = x; });
                if (String.IsNullOrEmpty(cmdArgs.Email))
                {
                    Console.Write("Enter username (email): ");
                    cmdArgs.Email = Console.ReadLine();
                }
                if (String.IsNullOrEmpty(cmdArgs.Password))
                {
                    Console.Write("Enter password        : ");
                    cmdArgs.Password = Console.ReadLine();
                }

                // Setup
                Console.Write("Setting up . . . . . ");
                CookieContainer cookieContainer;
                HttpClientHandler handler;
                HttpClient client;
                HttpResponseMessage response;
                try
                {
                    cookieContainer = new CookieContainer();
                    handler = new HttpClientHandler() { CookieContainer = cookieContainer };
                    client = new HttpClient(handler);
                    client.DefaultRequestHeaders.Add("Origin", "https://borderlands.com");
                    client.DefaultRequestHeaders.Add("Referer", "https://borderlands.com/en-US/vip/");
                }
                catch (Exception ex)
                {
                    Exit(ex);
                    return;
                }
                Console.WriteLine("success!");

                // Login
                Console.Write("Logging in as '" + cmdArgs.Email + "' . . . . . ");
                try
                {
                    var loginData = new Dictionary<string, string>
                    {
                        { "username", cmdArgs.Email },
                        { "password", cmdArgs.Password }
                    };
                    response = await client.PostAsJsonAsync("https://api.2k.com/borderlands/users/authenticate", loginData);
                    if (!response.StatusCode.Equals(HttpStatusCode.OK))
                    {
                        Console.WriteLine("failed! Did you use the correct password?");
                        Exit();
                        return;
                    }
                }
                catch (Exception ex)
                {
                    Exit(ex);
                    return;
                }
                Console.WriteLine("success!");

                // Get Previous Codes
                Console.Write("Getting previously redeemed codes . . . . . ");
                var codeTypes = new Dictionary<string, string>() 
                { 
                    {"vault", ""},
                    {"diamond", ""},
                    {"creator", ""},
                    {"email", ""},
                    {"boost", ""},
                };
                var redeemedCodes = new List<string>();
                try
                {
                    var redirectUrl = WebUtility.UrlDecode(response.Headers.First(x => x.Key.ToUpper() == "X-CT-REDIRECT").Value.ToArray()[0]);
                    await client.GetAsync(redirectUrl);
                    var previousCodesQuery = new 
                    {
                        model_data = new 
                        {
                            activity = new 
                            {
                                newest_activities = new 
                                {
                                    properties = new string [] { "notes", "title" },
                                    query = new 
                                    { 
                                        type = "user_activities_me", 
                                        args = new {row_start = 1, row_end = 1000000}
                                    }
                                }
                            }
                        }
                    };
                    response = await client.PostAsJsonAsync("https://2kgames.crowdtwist.com/request?widgetId=9470", previousCodesQuery);
                    var activities = JsonConvert.DeserializeObject<dynamic>(await response.Content.ReadAsStringAsync()).model_data.activity.newest_activities.ToObject<IEnumerable<dynamic>>();
                    foreach (var activity in activities)
                    {
                        var code = "";
                        var codeType = "";
                        try
                        {
                            code = activity.notes.ToObject<string>().ToUpper() ?? "";
                            codeType = activity.title.ToObject<string>().ToLower() ?? "";
                        }
                        catch {};
                        if (!String.IsNullOrEmpty(code) || !String.IsNullOrEmpty(codeType))
                        {
                            codeType = codeTypes.Keys.FirstOrDefault(x => codeType.Contains(x)) ?? "";
                            redeemedCodes.Add(codeType + ":" + code);
                        }
                    }
                }
                catch (Exception ex)
                {
                    Exit(ex);
                    return;
                }
                Console.WriteLine("success!");

                // Get New Codes
                Console.Write("Getting new codes . . . . . ");
                IBrowsingContext browsingContext;
                var codesToRedeem = new Dictionary<string, List<string>>()
                { 
                    {"vault", new List<string>()},
                    {"diamond", new List<string>()},
                    {"creator", new List<string>()},
                    {"email", new List<string>()},
                    {"boost", new List<string>()},
                };;
                try
                {
                    browsingContext = BrowsingContext.New(Configuration.Default.WithDefaultLoader());
                    var redditHtml =  await browsingContext.OpenAsync("https://www.reddit.com/r/borderlands3/comments/bxgq5p/borderlands_vip_program_codes/");
                    foreach (var row in redditHtml.QuerySelectorAll("tbody tr"))
                    {
                        var columns = row.QuerySelectorAll("td").Select(x => x.TextContent).ToArray();
                        var code = columns[0]?.ToUpper() ?? "";
                        var codeType = columns[3]?.ToLower() ?? "";
                        foreach (var key in codeTypes.Keys.Where(x => codeType.Contains(x)))
                        {
                            if (!redeemedCodes.Contains(key + ":" + code) && !columns[2].ToUpper().Contains("NO"))
                            {
                                codesToRedeem[key].Add(code);
                            }
                            
                        }
                        
                    }
                    if (codesToRedeem.Values.Sum(x => x.Count) == 0)
                    {
                        Console.WriteLine("No new codes at this time. Try again later.");
                        Exit();
                        return;
                    }
                }
                catch (Exception ex)
                {
                    Exit(ex);
                    return;
                }
                Console.WriteLine("success! Found " + codesToRedeem.Values.Sum(x => x.Count) + " unredeemed codes!");

                // Get Code URLs
                Console.Write("Getting code redemption URLs . . . . . ");
                IDocument widgetHtml;
                string widgetConfJson;
                dynamic widgetConf;
                try
                { 
                    widgetHtml = await browsingContext.OpenAsync("https://2kgames.crowdtwist.com/widgets/t/activity-list/9904/?__locale__=en#2");
                    widgetConfJson = widgetHtml.DocumentElement.QuerySelectorAll("script").First(x => x.TextContent.Contains("widgetConf")).TextContent;
                    widgetConf = JsonConvert.DeserializeObject<dynamic>(String.Join("", String.Join("", widgetConfJson.Split("widgetConf").Skip(1)).Split("=").Skip(1)).Split(';')[0].Trim());
                    foreach (var widget in widgetConf["entries"].ToObject<List<dynamic>>())
                    {
                        var codeType = codeTypes.Keys.FirstOrDefault(x => widget.link.widgetName.ToObject<string>().ToLower().Contains(x));
                        if (!String.IsNullOrEmpty(codeType))
                        {
                            codeTypes[codeType] = widget.link.widgetId.ToObject<string>().ToLower();
                        }
                    }
                }
                catch (Exception ex)
                {
                    Exit(ex);
                    return;
                }
                Console.WriteLine("success!");

                
                foreach(var keyValue in codesToRedeem)
                {
                    // Setup Code Type
                    if (keyValue.Value.Count < 1)
                        continue;
                    Console.WriteLine("Setting up codes of type '" + keyValue.Key + "' . . . . . ");
                    var codeRedemptionUrl = "";
                    try
                    {
                        var codeTypeUrl = "https://2kgames.crowdtwist.com/widgets/t/code-redemption/" + codeTypes[keyValue.Key] + "/";
                        widgetHtml = await browsingContext.OpenAsync(codeTypeUrl);
                        widgetConfJson = widgetHtml.DocumentElement.QuerySelectorAll("script").First(x => x.TextContent.Contains("widgetConf")).TextContent;
                        widgetConf = JsonConvert.DeserializeObject<dynamic>(String.Join("", String.Join("", widgetConfJson.Split("widgetConf").Skip(1)).Split("=").Skip(1)).Split(';')[0].Trim());
                        codeRedemptionUrl = "https://2kgames.crowdtwist.com/code-redemption-campaign/redeem?cid=" + widgetConf.campaignId.ToObject<string>();
                    }
                    catch
                    {
                        Console.WriteLine("failed! Unknown error.");
                        continue;
                    }
                    Console.WriteLine("success!");
                    
                    // Redeem Codes
                    foreach (var code in keyValue.Value)
                    {
                        Console.Write("Trying '" + keyValue.Key + "' code '" + code + "' . . . . . ");
                        try
                        {
                            response = await client.PostAsJsonAsync(codeRedemptionUrl, new { code = code });
                            var responseMessage = JsonConvert.DeserializeObject<dynamic>(await response.Content.ReadAsStringAsync());
                            try
                            {
                                if (response.StatusCode.Equals(HttpStatusCode.OK))
                                    Console.WriteLine(responseMessage.message.ToObject<string>());
                                else
                                    Console.WriteLine(responseMessage.exception.model.ToObject<string>());
                            }
                            catch
                            {
                                Console.WriteLine(response.StatusCode.ToString() + " . . . " + "(unable to read response effectively.)");
                            }
                        }
                        catch (Exception ex)
                        {
                            Exit(ex);
                            return;
                        }
                    }
                }
            }
            catch (Exception ex)
            {
                Exit(ex);
                return;
            }
            Console.WriteLine("Done!");
            Exit();
        }
    }
}
