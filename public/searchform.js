
const FilterSelect = cc('select');
const SearchInput = cc('input');
const SubmitBtn = cc('button');
const SearchForm = {
  view: () => m('form').addClass('SearchForm row g-1').append([
    m('div').addClass('col-auto').append(
      m(FilterSelect).addClass('form-select').append([
        m('option').text('tags').attr({value:'tags'}).prop('selected',true),
        m('option').text('name').attr({value:'title'}),
      ]),
    ),
    m('div').addClass('col input-group').append(
      m(SearchInput).addClass('form-control').attr({type:'text'}).prop({autofocus:true,required:true}),
      m(SubmitBtn).text('search').attr({type:'submit'}).addClass('btn btn-outline-primary').click(SearchForm.onsubmit),
    ),
  ]),
  onsubmit: (event) => {
    event.preventDefault();
    const pattern = $(SearchInput.id).val().trim();
    if (!pattern) {
      Alerts.insert('info', '请输入搜索内容');
      $(SearchInput.id).focus();
      return;
    }
    const searchBy = $(FilterSelect.id).val();
    if (searchBy == 'tags') SearchForm.searchTags();
    if (searchBy == 'title') SearchForm.searchTitle();
    if (searchBy == 'id') SearchForm.searchByID();
  },
  searchTags: () => {
    const tagSet = tagsStringToSet($(SearchInput.id).val());
    Alerts.clear();
    Alerts.insert('info', 'searching tags: ' + addPrefix(tagSet, '#'));
    const tags = Array.from(tagSet);
    const body = new FormData()
    body.append('tags', JSON.stringify(tags));
    body.append('file-type', 'image');
    const options = {method:'POST',url:'/api/search-tags',alerts:Alerts,buttonID:SubmitBtn.id,body:body};
    SearchForm.search(options);
  },
  searchTitle: () => {
    const pattern = $(SearchInput.id).val().trim();
    Alerts.clear();
    Alerts.insert('info', 'searching title: ' + pattern);
    const body = new FormData();
    body.append('pattern', pattern);
    body.append('file-type', 'image');
    const options = {method:'POST',url:'/api/search-title',alerts:Alerts,buttonID:SubmitBtn.id,body:body};
    SearchForm.search(options);
  },
  searchByID: () => {
    const fileid = $(SearchInput.id).val().trim();
    Alerts.clear();
    Alerts.insert('info', `查找文件 [id:${fileid}] 及与其重名的文件...`);
    const body = new FormData();
    body.append('id', fileid);
    body.append('file-type', 'image');
    const options = {method:'POST',url:'/api/search-by-id',alerts:Alerts,buttonID:SubmitBtn.id,body:body};
    SearchForm.search(options);
  },
  searchDamaged: () => {
    Alerts.clear();
    Alerts.insert('primary', '高级功能：查找数据库中的全部损坏文件...');
    const options = {method:'GET',url:'/api/search-damaged',alerts:Alerts,buttonID:SubmitBtn.id};
    SearchForm.search(options);
  },
  search: (options) => {
    ajax(options, SearchForm.onSuccess, SearchForm.onFail);
  },
  onSuccess: (files) => {
    if (!files || !files.length) {
      Alerts.insert('danger', '找不到相关文件');
      ImageList.clear();
      return;
    }
    Alerts.insert('success', `找到 ${files.length} 个文件`);
    ImageList.clear();
    ImageList.prepend(files);
  },
  onFail: () => {
    ImageList.clear();
  },
};