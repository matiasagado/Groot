from pystackapi import Site
from pystackapi.sites import StackOverflow

# site = Site(StackOverflow,  access_token="SlrSl3Ua*IqNkHVXp9MbGQ((", app_key="")
site = Site(StackOverflow)

info = site.get_info()

print(f'Total questions on StackOverflow: {info.total_questions}')

# serarch_results = site.search(intitle='nginx', tagged='nginx', page=1, pagesize=100)


# print(serarch_results)
# print(len(serarch_results))

# if __name__ == "__main__":

def fetch_all_results(site, intitle, tagged, pagesize, max_pages):
    page = 1
    while True:
        # search_results = site.search(intitle=intitle, tagged=tagged, page=page, pagesize=pagesize)
        search_results = site.advanced_search(q=intitle, page=page, pagesize=pagesize)
        num_results = len(search_results)
        
        # Process your results here (e.g., print them or save to a file)
        print(f"Page {page}: {num_results} items")

        print(search_results)

        for item in search_results: 
            

        # If less than 100 items are returned, it means we've reached the last page
        if num_results < pagesize or page >= max_pages:
            break
        
        # Increment the page number for the next iteration
        page += 1


# fetch_all_results(site, "nginx", "nginx", 1, 1)
fetch_all_results(site, "[nginx] OR [python]", "nginx", 1, 1)
