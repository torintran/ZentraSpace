const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { chromium } = require('playwright');

const app = process.env.ZS_BROWSER_BASE || 'http://localhost:8092';
const bank = process.env.ZS_BROWSER_MOCK || 'http://localhost:8093';
const envFile = process.env.ZS_BROWSER_ENV_FILE || path.resolve(__dirname, '../../.env');
const config = fs.readFileSync(envFile, 'utf8');
const password = config.match(/^ZS_OWNER_PASSWORD=([^\r\n]+)/m)?.[1];
const meterPassword = config.match(/^ZS_METER_PASSWORD=([^\r\n]+)/m)?.[1];
const period = '2026-11';
const roomName = 'PW-' + crypto.randomBytes(4).toString('hex');
const reference = roomName + '-PAY';

async function show(page, heading) {
  await page.getByRole('heading', {name: heading, exact: true}).waitFor();
}

async function fields(form, values) {
  for (const [name, value] of Object.entries(values)) {
    await form.locator(`[name="${name}"]`).fill(value);
  }
}

async function checkRoomOnboarding(browser, owner, tenant) {
  const name = roomName + '-TRONG';
  await owner.setViewportSize({width:390, height:844});
  await owner.goto(`${app}/?tab=rooms&period=${period}`);
  await owner.getByRole('link', {name:'Thêm phòng', exact:true}).click();
  await fields(owner.locator('form[action="/action/room"]'), {
    name, rent:'1800000', electricRate:'3000', waterRate:'10000', electric:'11', water:'4',
  });
  await owner.getByRole('button', {name:'Lưu phòng', exact:true}).click();
  await owner.locator('article').filter({hasText:name}).getByText('Phòng trống', {exact:true}).waitFor();
  await owner.getByRole('link', {name:`Xem chi tiết phòng ${name}`, exact:true}).click();
  const roomURL = owner.url();
  await owner.getByRole('navigation', {name:'Mục trong hồ sơ phòng'}).getByRole('link', {name:'Giá và phí'}).click();
  assert.equal(new URL(owner.url()).hash, '#room-fees');
  assert.ok(await owner.evaluate(() => document.documentElement.scrollWidth === innerWidth));
  await owner.getByText('Thêm khoản phí', {exact:true}).click();
  const fee = owner.locator('form[action="/action/room-fee"]');
  await fields(fee, {label:'Wi-Fi phòng trống', amount:'70000', start:period});
  await fee.getByRole('button', {name:'Lưu khoản phí'}).click();
  await owner.getByText('Wi-Fi phòng trống', {exact:true}).waitFor();
  await owner.goto(`${app}/?tab=meters&period=${period}`);
  assert.equal(await owner.locator('.meter-row').filter({hasText:name}).count(), 0);

  await owner.goto(roomURL);
  await owner.getByRole('link', {name:'Tiếp nhận người thuê', exact:true}).first().click();
  const lease = owner.locator('form[action="/action/lease"]');
  assert.equal(await lease.locator('[name="rent"]').inputValue(), '1800000');
  await fields(lease, {
    effective:period, tenant:'Người thuê mới', months:'12', phone:'0900000000', signedOn:'2026-11-01',
  });
  await lease.getByRole('button', {name:'Lưu hợp đồng'}).click();
  const contract = owner.locator('#room-contracts article').filter({hasText:'Người thuê mới'});
  await contract.getByText('Điện thoại 0900000000', {exact:true}).waitFor();
  await contract.getByText('Đính kèm tệp hợp đồng', {exact:true}).click();
  const upload = contract.locator('form[action$="/file"]');
  const filePath = await upload.getAttribute('action');
  const csrf = await upload.locator('[name="csrf"]').inputValue();
  const invalid = await owner.request.post(app + filePath, {
    headers:{Origin:app}, multipart:{csrf, document:{name:'wrong.pdf', mimeType:'application/pdf', buffer:Buffer.from('plain text')}},
  });
  assert.equal(invalid.status(), 422);
  const pdf = Buffer.from('%PDF-1.4\n% ZentraSpace synthetic lease document\n%%EOF\n');
  await upload.locator('[type="file"]').setInputFiles({name:'contract.pdf', mimeType:'application/pdf', buffer:pdf});
  await upload.getByRole('button', {name:'Lưu tệp hợp đồng'}).click();
  await contract.getByRole('link', {name:'Tải hợp đồng'}).waitFor();
  await contract.getByText('Email xác minh bảng kê', {exact:true}).click();
  const contact = contract.locator('form[action="/action/lease-email"]');
  await fields(contact, {email:'tenant@example.test'});
  await contact.getByRole('button', {name:'Lưu email'}).click();
  await owner.getByRole('status').filter({hasText:'Đã lưu email xác minh'}).waitFor();
  await contract.getByText('Email xác minh bảng kê', {exact:true}).click();
  assert.equal(await contact.locator('input[name="email"]').inputValue(), 'tenant@example.test');
  const file = await owner.request.get(app + filePath);
  assert.equal(file.status(), 200);
  assert.match(file.headers()['content-disposition'], /^attachment;/);
  assert.equal(file.headers()['cache-control'], 'no-store');
  assert.ok((await file.body()).equals(pdf));
  assert.equal((await tenant.request.get(app + filePath, {maxRedirects:0})).status(), 303);
  for (const { username, credential, status } of [
    {username:'meter', credential:meterPassword, status:403},
    ...JSON.parse(process.env.ZS_STAFF_ACCOUNTS || '[]')
      .filter(account => account.Username === 'owner-secondary')
      .map(account => ({username:account.Username, credential:account.Password, status:404})),
  ]) {
    const session = await browser.newContext();
    try {
      const signIn = await session.request.post(app + '/login', {
        headers:{Origin:app}, form:{account:username, password:credential}, maxRedirects:0,
      });
      assert.equal(signIn.status(), 303);
      assert.equal((await session.request.get(app + filePath)).status(), status);
    } finally {
      await session.close();
    }
  }

  await owner.goto(`${app}/?tab=meters&period=${period}`);
  const row = owner.locator('.meter-row').filter({hasText:name});
  await fields(row, {electric:'12', water:'5'});
  await row.getByRole('button', {name:'Lưu chỉ số'}).click();
  await owner.goto(`${app}/?tab=billing&period=${period}`);
  await owner.getByRole('button', {name:'Tạo bảng kê nháp'}).click();
  await owner.locator('article').filter({hasText:name}).getByRole('link', {name:'Xem bảng kê'}).click();
  await owner.getByText('Wi-Fi phòng trống', {exact:true}).waitFor();
  await owner.getByText('1.883.000 ₫', {exact:true}).waitFor();
  console.log('PASS Playwright: prepare room and fees, start first lease, private document, then bill');
}

async function checkPropertyWorkspace(browser, owner, tenant) {
  await owner.setViewportSize({width:1440, height:900});
  await owner.goto(`${app}/?period=${period}`);
  const sidebar = owner.getByRole('navigation', {name:'Khu vực quản lý'});
  const header = owner.getByRole('banner');
  const desktop = owner.locator('.workspace-sidebar-desktop');
  const expandedWidth = (await desktop.boundingBox()).width;
  const collapse = header.getByRole('button', {name:'Thu gọn thanh bên'});
  const togglePosition = await collapse.boundingBox();
  await collapse.press('Enter');
  const expand = header.getByRole('button', {name:'Mở rộng thanh bên'});
  await expand.waitFor();
  assert.deepEqual(await expand.boundingBox(), togglePosition);
  assert.equal(await sidebar.getByRole('button').count(), 0);
  assert.equal(await expand.getAttribute('aria-expanded'), 'false');
  assert.ok((await desktop.boundingBox()).width < expandedWidth);
  await sidebar.getByRole('link', {name:'Dãy trọ', exact:true}).click();
  await expand.waitFor();
  await owner.reload();
  await expand.waitFor();
  await expand.click();
  assert.equal(await owner.getByRole('button', {name:'Thu gọn thanh bên'}).getAttribute('aria-expanded'), 'true');
  assert.equal((await desktop.boundingBox()).width, expandedWidth);
  console.log('PASS Playwright: collapsible desktop sidebar remembers the setting across navigation and reload');
  const siteName = roomName + '-DAY';
  const site = owner.locator('form[action="/action/site"]');
  await fields(site, {name:siteName, address:'Địa chỉ mô phỏng', capacity:'3'});
  const png = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGNwqk8FAAItASeOCvhFAAAAAElFTkSuQmCC','base64');
  await owner.route('https://tile.openstreetmap.org/**', route =>
    route.fulfill({status:200, contentType:'image/png', body:png}));
  await owner.context().grantPermissions(['geolocation'], {origin:app});
  await owner.context().setGeolocation({latitude:10.7769, longitude:106.7009, accuracy:15});
  await site.getByText('Chọn vị trí trên bản đồ', {exact:true}).click();
  const map = site.getByRole('region', {name:'Bản đồ chọn vị trí dãy trọ'});
  await map.locator('img.leaflet-tile-loaded').first().waitFor();
  const locate = site.getByRole('button', {name:'Dùng vị trí hiện tại'});
  await locate.click();
  await site.getByText('Vị trí đã chọn: 10.776900, 106.700900', {exact:true}).waitFor();
  await map.click({position:{x:45, y:75}});
  assert.notEqual(await site.locator('[name="coordinates"]').inputValue(), '10.776900, 106.700900');
  await locate.click();
  await site.getByText('Vị trí đã chọn: 10.776900, 106.700900', {exact:true}).waitFor();
  await site.getByRole('button', {name:'Lưu dãy trọ'}).click();
  const siteRow = owner.locator('article').filter({hasText:siteName});
  const sitePanel = owner.locator('#site-map-panel');
  const siteDetails = owner.locator('#site-details-panel');
  await sitePanel.getByRole('region', {name:'Bản đồ vị trí đã lưu'}).locator('svg path').first().waitFor();
  assert.match(await sitePanel.getByRole('link', {name:'Mở OpenStreetMap'}).getAttribute('href'), /openstreetmap\.org.*mlat=10\.776900/);
  assert.equal(await sitePanel.getByText('Xóa dãy trọ', {exact:true}).count(), 0);
  const editBounds = await sitePanel.getByRole('link', {name:'Sửa vị trí', exact:true}).boundingBox();
  const mapLinkBounds = await sitePanel.getByRole('link', {name:'Mở OpenStreetMap'}).boundingBox();
  assert.equal(editBounds.height, 46);
  assert.equal(mapLinkBounds.height, editBounds.height);
  assert.equal(mapLinkBounds.y, editBounds.y);
  const addBounds = await owner.getByRole('link', {name:'Thêm dãy trọ'}).boundingBox();
  const rowBounds = await siteRow.boundingBox();
  assert.ok(rowBounds.y - addBounds.y - addBounds.height >= 12);
  await owner.setViewportSize({width:390, height:844});
  await sitePanel.getByRole('link', {name:'Sửa vị trí', exact:true}).click();
  const updateLocation = sitePanel.locator('form[action="/action/site-location"]');
  await updateLocation.getByRole('region', {name:'Bản đồ chọn vị trí dãy trọ'}).locator('svg path').first().waitFor();
  await updateLocation.getByText('Vị trí đã chọn: 10.776900,106.700900', {exact:true}).waitFor();
  await owner.context().setGeolocation({latitude:10.7801, longitude:106.7012, accuracy:15});
  await updateLocation.getByRole('button', {name:'Dùng vị trí hiện tại'}).click();
  await updateLocation.getByText('Vị trí đã chọn: 10.780100, 106.701200', {exact:true}).waitFor();
  assert.equal(await owner.evaluate(() => document.documentElement.scrollWidth), 390);
  await updateLocation.getByRole('button', {name:'Lưu vị trí'}).click();
  assert.match(await sitePanel.getByRole('link', {name:'Mở OpenStreetMap'}).getAttribute('href'), /mlat=10\.780100/);
  await siteRow.getByRole('link', {name:'Chi tiết', exact:true}).click();
  await siteDetails.getByRole('heading', {name:'Chi tiết dãy trọ'}).waitFor();
  await siteDetails.getByText('Địa chỉ mô phỏng', {exact:true}).waitFor();
  await siteDetails.getByText('0/3', {exact:true}).waitFor();
  assert.equal(await owner.evaluate(() => document.documentElement.scrollWidth), 390);
  await owner.setViewportSize({width:1440, height:900});
  console.log('PASS Playwright: site details, aligned actions, map, and editing a saved location on mobile');

  await sidebar.getByRole('link', {name:'Mẫu giá và phí', exact:true}).click();
  const policyName = roomName + '-MAU';
  const policy = owner.locator('form[action="/action/policy"]');
  await fields(policy, {name:policyName, rent:'1200000', electricRate:'3000', waterRate:'10000', fee:'50000', feeLabel:'Wi-Fi theo mẫu'});
  await policy.getByRole('button', {name:'Lưu mẫu giá và phí'}).click();
  await owner.locator('article').getByText(policyName, {exact:true}).waitFor();
  await sidebar.getByRole('link', {name:'Dãy trọ', exact:true}).click();
  await siteRow.getByRole('link', {name:'Xem phòng'}).click();
  await owner.getByRole('link', {name:'Thêm phòng', exact:true}).click();
  const room = owner.locator('form[action="/action/room"]');
  const name = roomName + '-P01';
  const siteID = await room.locator('[name="site"]').inputValue();
  assert.ok(siteID);
  await fields(room, {name, description:'Phòng có cửa sổ'});
  await room.getByRole('combobox', {name:'Mẫu giá và phí'}).selectOption({label:policyName});
  const policyID = await room.locator('[name="policy"]').inputValue();
  const csrf = await room.locator('[name="csrf"]').inputValue();
  await room.getByRole('button', {name:'Lưu phòng', exact:true}).click();
  const roomCard = owner.locator('article').filter({hasText:name});
  assert.equal(await roomCard.getAttribute('data-room-state'), 'vacant');
  for (const suffix of ['P02', 'P03']) {
    const extra = await owner.request.post(app + '/action/room', {
      headers:{Origin:app}, form:{csrf, period, name:roomName+'-'+suffix, site:siteID, rent:'700000',
        electricRate:'1000', waterRate:'5000', electric:'0', water:'0'}, maxRedirects:0,
    });
    assert.equal(extra.status(), 303);
  }
  const tooManyRooms = await owner.request.post(app + '/action/room', {
    headers:{Origin:app}, form:{csrf, period, name:name+'-extra', site:siteID, policy:policyID, electric:'0', water:'0'},
  });
  assert.equal(tooManyRooms.status(), 422);

  await roomCard.getByRole('link', {name:`Xem chi tiết phòng ${name}`, exact:true}).click();
  const roomURL = owner.url();
  await owner.getByText('Phòng có cửa sổ', {exact:true}).waitFor();
  await owner.getByText('Thêm ảnh phòng', {exact:true}).click();
  const photoForm = owner.locator('form[action$="/photo"]');
  const photoPath = await photoForm.getAttribute('action');
  await photoForm.locator('[type="file"]').setInputFiles({name:'room.png', mimeType:'image/png', buffer:png});
  await photoForm.getByRole('button', {name:'Lưu ảnh phòng'}).click();
  await owner.getByRole('img', {name:`Ảnh phòng ${name}`}).waitFor();
  assert.equal((await owner.request.get(app + photoPath)).headers()['content-type'], 'image/jpeg');
  assert.equal((await tenant.request.get(app + photoPath, {maxRedirects:0})).status(), 303);

  await owner.getByRole('link', {name:'Tiếp nhận người thuê', exact:true}).first().click();
  const lease = owner.locator('form[action="/action/lease"]');
  assert.equal(await lease.locator('[name="rent"]').inputValue(), '1200000');
  await fields(lease, {effective:period, tenant:'Người đại diện dãy mẫu'});
  await lease.getByRole('button', {name:'Lưu hợp đồng'}).click();
  for (const person of ['Người ở cùng 01','Người ở cùng 02']) {
    await owner.getByText('Thêm người ở cùng', {exact:true}).click();
    const resident = owner.locator('form[action="/action/resident"]');
    await fields(resident, {name:person, joinedOn:'2026-11-01'});
    await resident.getByRole('button', {name:'Lưu người ở'}).click();
    await owner.locator('#room-residents').getByText(person, {exact:true}).waitFor();
  }
  await owner.goto(`${app}/?tab=rooms&period=${period}&site=${siteID}`);
  assert.equal(await roomCard.getAttribute('data-room-state'), 'occupied');
  await roomCard.getByRole('img', {name:'3 người ở', exact:true}).waitFor();
  assert.equal(await roomCard.locator('.person-avatar').count(), 3);

  await owner.goto(`${app}/?tab=meters&period=${period}`);
  const meter = owner.locator('.meter-row').filter({hasText:name});
  await fields(meter, {electric:'1', water:'1'});
  await meter.getByRole('button', {name:'Lưu chỉ số'}).click();
  await owner.goto(`${app}/?tab=billing&period=${period}`);
  await owner.getByRole('button', {name:'Tạo bảng kê nháp'}).click();
  await owner.locator('article').filter({hasText:name}).getByRole('link', {name:'Xem bảng kê'}).click();
  await owner.getByText('1.263.000 ₫', {exact:true}).waitFor();
  await owner.getByRole('button', {name:'Phát hành bảng kê'}).click();
  const payment = owner.locator('form[action="/action/post"]').first();
  await fields(payment, {amount:'90000'});
  await payment.locator('[name="method"]').selectOption('cash');
  await payment.getByRole('button', {name:'Ghi khoản thu'}).click();
  await owner.goto(roomURL);
  const timeline = owner.locator('#room-timeline');
  await timeline.getByText('Hợp đồng có hiệu lực', {exact:true}).waitFor();
  await timeline.getByText('Thu tiền phòng', {exact:true}).waitFor();
  await timeline.getByText('90.000 ₫', {exact:true}).waitFor();

  await owner.goto(`${app}/?tab=policies&period=${period}`);
  const batch = owner.locator('form[action="/action/policy-preview"]');
  await batch.locator('[name="scope"]').selectOption({label:siteName});
  await batch.locator('[name="policy"]').selectOption({label:policyName});
  await fields(batch, {effective:'2026-12'});
  await batch.getByRole('button', {name:'Xem trước thay đổi'}).click();
  await owner.getByText('2 phòng đủ điều kiện', {exact:true}).waitFor();
  await owner.locator('article').filter({hasText:name}).getByText('Bỏ qua · Phòng đã có hồ sơ hợp đồng').waitFor();
  await owner.getByRole('button', {name:'Áp dụng cho 2 phòng'}).click();
  await owner.getByRole('status').filter({hasText:'Đã lên lịch áp dụng mẫu giá và phí'}).waitFor();
  await owner.goto(`${app}/?tab=rooms&period=${period}&site=${siteID}`);
  await owner.getByRole('link', {name:`Xem chi tiết phòng ${roomName}-P02`, exact:true}).click();
  await owner.locator('#room-timeline').getByText('Đã lên lịch áp dụng mẫu giá').waitFor();
  await owner.getByRole('link', {name:'Tiếp nhận người thuê', exact:true}).first().click();
  const scheduledLease = owner.locator('form[action="/action/lease"]');
  const tariff = scheduledLease.getByRole('checkbox', {name:'Dùng giá theo kỳ hiệu lực đã chọn'});
  assert.ok(await tariff.isChecked());
  await fields(scheduledLease, {effective:'2026-12', tenant:'Người thuê dùng mẫu'});
  await scheduledLease.getByRole('button', {name:'Lưu hợp đồng'}).click();
  await owner.locator('#room-profile').getByText('1.200.000 ₫', {exact:true}).waitFor();
  console.log('PASS Playwright: preview a dated policy batch, exclude the occupied room, and inherit prices on move-in');

  const protectedRoom = await owner.request.post(app + '/action/room-delete', {
    headers:{Origin:app}, form:{csrf, period, id:new URL(roomURL).pathname.split('/').at(-1), confirm:'yes'},
  });
  assert.equal(protectedRoom.status(), 422);
  const protectedSite = await owner.request.post(app + '/action/site-delete', {
    headers:{Origin:app}, form:{csrf, period, id:siteID, confirm:'yes'},
  });
  assert.equal(protectedSite.status(), 422);

  await owner.goto(`${app}/?tab=rooms&period=${period}&site=${siteID}`);
  await owner.getByRole('link', {name:`Xem chi tiết phòng ${roomName}-P03`, exact:true}).click();
  const unusedRoomURL = owner.url();
  await owner.getByText('Thêm ảnh phòng', {exact:true}).click();
  const unusedPhoto = owner.locator('form[action$="/photo"]');
  const unusedPhotoPath = await unusedPhoto.getAttribute('action');
  await unusedPhoto.locator('input[type=file]').setInputFiles({name:'unused.png', mimeType:'image/png', buffer:png});
  await unusedPhoto.getByRole('button', {name:'Lưu ảnh phòng'}).click();
  assert.equal((await owner.request.get(app + unusedPhotoPath)).status(), 200);
  await owner.locator('.property-delete summary').click();
  const deleteRoom = owner.locator('form[action="/action/room-delete"]');
  await deleteRoom.getByRole('checkbox').check();
  await deleteRoom.getByRole('button', {name:'Xóa phòng', exact:true}).click();
  await owner.getByRole('status').filter({hasText:'Đã xóa phòng'}).waitFor();
  assert.equal((await owner.request.get(unusedRoomURL)).status(), 404);
  assert.equal((await owner.request.get(app + unusedPhotoPath)).status(), 404);

  await owner.goto(`${app}/?tab=sites&period=${period}`);
  const emptySiteName = siteName + '-TRONG';
  await fields(site, {name:emptySiteName, capacity:'1'});
  await site.getByRole('button', {name:'Lưu dãy trọ'}).click();
  await sitePanel.getByText('Chưa ghim vị trí.', {exact:false}).waitFor();
  await sitePanel.getByRole('link', {name:'Sửa vị trí', exact:true}).click();
  await sitePanel.getByRole('region', {name:'Bản đồ chọn vị trí dãy trọ'}).locator('img.leaflet-tile-loaded').first().waitFor();
  await sitePanel.getByRole('link', {name:'Hủy', exact:true}).click();
  await sitePanel.getByRole('link', {name:'Chi tiết dãy trọ', exact:true}).click();
  await siteDetails.getByRole('link', {name:'Sửa thông tin', exact:true}).click();
  const siteEdit = siteDetails.locator('form[action="/action/site-edit"]');
  const updatedSiteName = emptySiteName + '-MOI';
  await fields(siteEdit, {name:updatedSiteName, capacity:'2'});
  await siteEdit.getByRole('button', {name:'Lưu thông tin'}).click();
  assert.equal(await siteDetails.locator('.site-details-name').textContent(), updatedSiteName);
  await siteDetails.getByText('0/2', {exact:true}).waitFor();
  await siteDetails.locator('.property-delete summary').click();
  const deleteSite = siteDetails.locator('form[action="/action/site-delete"]');
  await deleteSite.getByRole('checkbox').check();
  await deleteSite.getByRole('button', {name:'Xóa dãy trọ', exact:true}).click();
  await owner.getByRole('status').filter({hasText:'Đã xóa dãy trọ'}).waitFor();
  assert.equal(await owner.locator('article').filter({hasText:updatedSiteName}).count(), 0);
  console.log('PASS Playwright: delete an unused room and photo, delete an empty site, protect existing history');

  await owner.goto(roomURL);
  await owner.getByRole('link', {name:'Trả phòng và tất toán', exact:true}).click();
  const todayVN = new Intl.DateTimeFormat('sv-SE', {timeZone:'Asia/Ho_Chi_Minh', year:'numeric', month:'2-digit', day:'2-digit'}).format(new Date());
  assert.equal(await owner.locator('input[name="date"]').inputValue(), todayVN);
  assert.equal(await owner.locator('input[name="date"]').getAttribute('max'), todayVN);
  await owner.locator('input[name="date"]').fill('2026-11-15');
  await owner.getByRole('button', {name:'Xem trước tất toán'}).click();
  const checkout = owner.locator('form[action="/action/checkout"]');
  await checkout.locator('textarea[name="note"]').fill('Đã kiểm tra bàn giao');
  await checkout.getByRole('checkbox').check();
  await checkout.getByRole('button', {name:'Xác nhận trả phòng'}).click();
  await owner.getByRole('status').filter({hasText:'Đã trả phòng và lưu tất toán'}).waitFor();
  await owner.getByRole('heading', {name:'Phòng đã trả và đang trống'}).waitFor();
  assert.equal(await owner.locator('#room-residents article').count(), 0);
  await owner.locator('#room-timeline').getByText('Trả phòng và tất toán', {exact:true}).waitFor();
  await owner.goto(`${app}/?tab=rooms&period=${period}&site=${siteID}`);
  assert.equal(await roomCard.getAttribute('data-room-state'), 'vacant');
  assert.equal(await roomCard.locator('.person-avatar').count(), 0);
  await owner.goto(`${app}/?tab=rooms&period=2026-12&site=${siteID}`);
  assert.equal(await roomCard.getAttribute('data-room-state'), 'vacant');
  await sidebar.getByRole('link', {name:'Cấu hình', exact:true}).click();
  await owner.getByRole('heading', {name:'Email xác minh qua SMTP'}).waitFor();
  await owner.getByText('Chưa bật', {exact:true}).waitFor();
  const smtpForm = owner.locator('form[action="/action/smtp-save"]');
  const smtpSecret = 'only-for-isolated-browser-test';
  await fields(smtpForm, {host:'smtp.example.test', port:'587', sender:'sender@example.test', username:'smtp-user', password:smtpSecret});
  await smtpForm.locator('[name=mode]').selectOption('starttls');
  await smtpForm.getByRole('button', {name:'Lưu và bật SMTP'}).click();
  await owner.getByRole('status').filter({hasText:'Đã lưu SMTP'}).waitFor();
  await owner.reload();
  assert.equal(await smtpForm.locator('[name=host]').inputValue(), 'smtp.example.test');
  assert.equal(await smtpForm.locator('[name=password]').inputValue(), '');
  assert.ok(!(await owner.content()).includes(smtpSecret));
  const secondary = JSON.parse(process.env.ZS_STAFF_ACCOUNTS || '[]').find(account => account.Username === 'owner-secondary');
  if (secondary) {
    const context = await browser.newContext();
    try {
      const page = await context.newPage();
      await page.request.post(`${app}/login`, {headers:{Origin:app}, form:{account:secondary.Username,password:secondary.Password}});
      await page.goto(`${app}/?tab=settings`);
      await page.getByText('Đã bật', {exact:true}).waitFor();
      assert.equal(await page.locator('#smtp input').count(), 0);
      assert.ok(!(await page.content()).includes('smtp.example.test'));
      const secondaryCSRF = await page.locator('input[name=csrf]').first().inputValue();
      assert.equal((await page.request.post(`${app}/action/smtp-disable`, {headers:{Origin:app}, form:{csrf:secondaryCSRF}})).status(), 403);
    } finally { await context.close(); }
  }
  await owner.getByRole('button', {name:'Tắt SMTP', exact:true}).click();
  await owner.getByText('Chưa bật', {exact:true}).waitFor();
  console.log('PASS Playwright: immediate vacancy, saved history and global SMTP administration');

  await owner.setViewportSize({width:320, height:820});
  await header.getByRole('button', {name:'Thu gọn thanh bên'}).waitFor({state:'hidden'});
  const mobileMenu = owner.locator('.workspace-sidebar-mobile');
  await mobileMenu.locator('summary').click();
  await mobileMenu.getByRole('link', {name:'Đối soát ngân hàng', exact:true}).click();
  await owner.getByText('Đổi tài khoản', {exact:true}).click();
  const banks = owner.getByRole('combobox', {name:'Ngân hàng', exact:true});
  assert.ok(await banks.locator('option').count() > 50);
  assert.equal(await banks.locator('option').filter({hasText:'Vietcombank'}).count(), 1);
  assert.equal(await owner.locator('[name="bin"]').count(), 0);
  assert.ok(await owner.evaluate(() => document.documentElement.scrollWidth === innerWidth));
  console.log('PASS Playwright: sidebar, site and map, room policy, private photo, three residents, timeline and bank names');
}

async function run() {
  assert.ok(password, 'The test environment requires an owner password.');
  const channel = process.env.ZS_BROWSER_CHANNEL;
  const browser = await chromium.launch(channel ? {channel} : {});
  try {
    const owner = await browser.newPage();
    const clientErrors = [];
    owner.on('pageerror', error => clientErrors.push(error.message));
    await owner.goto(app + '/login');
    assert.equal(await owner.locator('datalist').count(), 0);
    const loginProfile = owner.getByRole('combobox', {name:'Tài khoản', exact:true});
    assert.equal(await loginProfile.evaluate(el => getComputedStyle(el).appearance), 'none');
    assert.equal(await owner.locator('.field-select .lucide-chevron-down').count(), 1);
    await loginProfile.selectOption('other');
    await owner.getByRole('textbox', {name:'Tên đăng nhập'}).waitFor();
    await loginProfile.selectOption('meter');
    await owner.getByRole('textbox', {name:'Tên đăng nhập'}).waitFor({state:'hidden'});
    await loginProfile.selectOption('owner');
    const passwordInput = owner.getByLabel(/^Mật khẩu/);
    const togglePassword = owner.getByRole('button', {name:'Hiện hoặc ẩn mật khẩu'});
    await passwordInput.fill(password);
    await togglePassword.click();
    await owner.waitForFunction(() => document.querySelector('[name="password"]').type === 'text');
    await togglePassword.click();
    assert.equal(await passwordInput.getAttribute('type'), 'password');
    await owner.getByRole('button', {name:'Đăng nhập', exact:true}).click();
    await show(owner, 'Tổng quan kỳ thu');
    await owner.goto(`${app}/?tab=rooms&period=${period}`);
    if (process.env.ZS_BROWSER_ENV_FILE) {
      const demoRooms = owner.locator('article').filter({hasText:'HD-MAU-2026-11-'});
      assert.equal(await demoRooms.count(), 6);
      const search = owner.getByRole('searchbox', {name:'Tìm theo mã phòng'});
      await search.fill('HD-MAU-2026-11-P04');
      await owner.getByRole('link', {name:'Xem chi tiết phòng HD-MAU-2026-11-P05', exact:true}).waitFor({state:'hidden'});
      assert.equal(await demoRooms.count(), 1);
      await search.clear();
      assert.ok((await demoRooms.filter({hasText:'P04'}).textContent()).includes('18 tháng'));
      await owner.getByRole('link', {name:'Xem chi tiết phòng HD-MAU-2026-11-P05', exact:true}).click();
      await owner.getByText('Từ 2026-11 đến hết 2027-10', {exact:true}).waitFor();
      assert.equal(await owner.locator('#room-contracts article').count(), 1);
      await owner.getByRole('link', {name:'Xem tất cả hợp đồng'}).click();
      await owner.getByText('Từ 2026-05 đến hết 2026-10', {exact:true}).waitFor();
      await owner.getByText('Khách thuê mẫu 05 · Ký lại', {exact:true}).waitFor();
      await owner.locator('[name="archivePeriod"]').fill('2026-05');
      await owner.getByRole('button', {name:'Lọc lịch sử'}).click();
      assert.equal(await owner.locator('#room-contracts article').count(), 1);
      await owner.goto(`${app}/?tab=rooms&period=${period}`);
      console.log('PASS Playwright: six stable contract fixtures, eighteen-month term and renewal history');
    }
    await owner.getByRole('link', {name:'Thêm phòng', exact:true}).click();
    await fields(owner.locator('form[action="/action/room"]'), {
      name:roomName, tenant:'Người thuê mô phỏng', rent:'2000000',
      electricRate:'3000', waterRate:'10000', electric:'0', water:'0',
    });
    await owner.getByRole('button', {name:'Lưu phòng', exact:true}).click();
    await owner.getByRole('link', {name:`Xem chi tiết phòng ${roomName}`, exact:true}).click();
    await show(owner, roomName);
    const roomURL = owner.url();
    assert.deepEqual(clientErrors, [], 'The Next.js UI hydrates without runtime errors');
    for (const data of [
      {label:'Wi-Fi', amount:'80000', kind:'monthly'},
      {label:'Phụ thu vệ sinh', amount:'35000', kind:'once'},
    ]) {
      await owner.getByText('Thêm khoản phí', {exact:true}).click();
      const form = owner.locator('form[action="/action/room-fee"]');
      await fields(form, {label:data.label, amount:data.amount, start:period});
      await form.locator('[name="kind"]').selectOption(data.kind);
      await form.getByRole('button', {name:'Lưu khoản phí'}).click();
      await owner.getByRole('status').filter({hasText:'Đã lưu khoản phí'}).waitFor();
    }
    await owner.goto(`${app}/?tab=meters&period=${period}`);
    const meter = owner.locator('.meter-row').filter({hasText:roomName});
    await fields(meter, {electric:'10', water:'2'});
    await meter.getByRole('button', {name:'Lưu chỉ số'}).click();
    await owner.goto(`${app}/?tab=billing&period=${period}`);
    await owner.getByRole('button', {name:'Tạo bảng kê nháp'}).click();
    await owner.locator('article').filter({hasText:roomName}).getByRole('link', {name:'Xem bảng kê'}).click();
    await owner.getByText('2.165.000 ₫', {exact:true}).waitFor();
    const invoiceURL = owner.url();
    await owner.getByRole('button', {name:'Phát hành bảng kê'}).click();
    await owner.getByText('Đã phát hành', {exact:true}).waitFor();

    await owner.goto(roomURL);
    await owner.getByText('Hủy phí Phụ thu vệ sinh', {exact:true}).click();
    const refused = owner.waitForResponse(r => r.request().method() === 'POST' && r.url().endsWith('/action/room-fee-stop'));
    await owner.getByRole('button', {name:'Xác nhận hủy'}).click();
    assert.equal((await refused).status(), 422, 'An issued charge stays on its invoice.');

    await owner.goto(invoiceURL);
    await owner.getByRole('button', {name:'Tạo link mới'}).click();
    const code = await owner.locator('.code strong').textContent();
    const message = await owner.locator('#share-message').inputValue();
    const invite = message.match(/https?:\/\/[^\s]+/)?.[0];
    assert.ok(invite, 'The invitation contains a link.');
    const tenant = await browser.newPage({viewport:{width:390,height:844}, isMobile:true, hasTouch:true});
    await tenant.goto(invite);
    await show(tenant, 'Xác minh để xem bảng kê');
    assert.equal(await tenant.getByText('Wi-Fi', {exact:true}).count(), 0);
    assert.equal((await tenant.request.get(invite + '/qr')).status(), 404);
    await tenant.getByLabel('Mã xác minh').fill(code.trim());
    await tenant.getByRole('button', {name:'Xem bảng kê'}).click();
    await show(tenant, 'Bảng kê tiền thuê');
    await tenant.getByText('Wi-Fi', {exact:true}).waitFor();
    await tenant.getByText('Phụ thu vệ sinh', {exact:true}).waitFor();
    assert.equal(await tenant.locator('.tenant-balance').textContent(), '2.165.000 ₫');
    assert.ok(await tenant.evaluate(() => document.documentElement.scrollWidth === document.documentElement.clientWidth));
    await tenant.getByRole('link', {name:'Xem mã thanh toán'}).click();
    const memo = await tenant.locator('dl').filter({hasText:'Nội dung chuyển khoản'})
      .locator('div').filter({hasText:'Nội dung chuyển khoản'}).locator('dd').textContent();
    assert.equal((await tenant.request.get(invite + '/qr')).status(), 200);

    const mock = await browser.newPage();
    await mock.goto(bank);
    await fields(mock.locator('form'), {reference, date:'2026-11-26', amount:'165000', memo:memo.trim()});
    await mock.getByRole('button', {name:'Gửi giao dịch'}).click();
    await mock.getByRole('status').filter({hasText:'Đã nhận giao dịch'}).waitFor();
    await owner.goto(app + '/bank');
    await owner.locator('article').filter({hasText:reference}).getByRole('link', {name:'Duyệt giao dịch'}).click();
    await owner.getByLabel('Cách ghi nhận').selectOption('new');
    await owner.getByRole('checkbox', {name:'Tôi đã kiểm tra bảng kê và khoản thu.'}).check();
    await owner.getByRole('button', {name:'Xác nhận đối soát'}).click();
    await tenant.reload();
    assert.equal(await tenant.locator('.tenant-balance').textContent(), '2.000.000 ₫');
    await tenant.getByText('Đã thu ròng', {exact:true}).waitFor();
    console.log('PASS Playwright: room fees → issued bill → mobile tenant verification → mock payment → updated balance');

    const following = '2026-12';
    await owner.goto(`${app}/?tab=meters&period=${following}`);
    const nextMeter = owner.locator('.meter-row').filter({hasText:roomName});
    await fields(nextMeter, {electric:'12', water:'3'});
    await nextMeter.getByRole('button', {name:'Lưu chỉ số'}).click();
    await owner.goto(`${app}/?tab=billing&period=${following}`);
    await owner.getByRole('button', {name:'Tạo bảng kê nháp'}).click();
    await owner.locator('article').filter({hasText:roomName}).getByRole('link', {name:'Xem bảng kê'}).click();
    const nextInvoiceURL = owner.url();
    await owner.getByRole('button', {name:'Phát hành bảng kê'}).click();
    await owner.getByText('Đã phát hành', {exact:true}).waitFor();

    await owner.goto(`${app}/?tab=dashboard&period=${following}`);
    await show(owner, 'Tổng quan kỳ thu');
    const tile = owner.locator('.room-map a').filter({hasText:roomName});
    assert.ok((await tile.textContent()).includes('Còn phải thu'));
    assert.equal(new URL(await tile.getAttribute('href'), app).href, nextInvoiceURL);
    assert.ok(Number(await owner.locator('#issue-progress').getAttribute('value')) >= 1);
    const finance = owner.locator('.dashboard-finances');
    assert.equal(await finance.getAttribute('open'), null);
    await finance.locator('summary').click();
    await show(owner, 'Đã thu và còn phải thu trong 6 kỳ');
    assert.equal(await finance.locator('.trend-month').count(), 6);
    assert.ok(Number(await finance.getByLabel('Đã thu ròng kỳ 11/2026', {exact:true}).getAttribute('value')) >= 165000);
    assert.ok(Number(await finance.getByLabel('Còn phải thu kỳ 12/2026', {exact:true}).getAttribute('value')) > 0);
    for (const width of [1440, 390, 320]) {
      await owner.setViewportSize({width, height:900});
      assert.ok(await owner.evaluate(() => document.documentElement.scrollWidth === document.documentElement.clientWidth),
        `Dashboard has no horizontal overflow at ${width}px`);
    }
    await tile.click();
    assert.equal(owner.url(), nextInvoiceURL);
    console.log('PASS Playwright: room map deep link, progress bars, six-period collections and responsive dashboard');

    await owner.goto(`${app}/cash?period=${period}`);
    await show(owner, 'Thu chi và cọc');
    const cost = owner.getByRole('form', {name:'Ghi chi phí'});
    await fields(cost, {date:'2026-11-26', amount:'120000', reference:'PW-CHI', note:'Thay bóng đèn'});
    await cost.getByRole('button', {name:'Lưu khoản chi'}).click();
    await owner.getByRole('status').filter({hasText:'Đã ghi giao dịch'}).waitFor();
    const deposit = owner.getByRole('form', {name:'Nhận hoặc hoàn cọc'});
    const roomValue = await deposit.locator('option').filter({hasText:roomName}).getAttribute('value');
    for (const [kind, amount, receipt, note] of [
      ['deposit_in','1000000','PW-COC','Cọc đầu kỳ'],
      ['deposit_out','400000','PW-HOAN','Hoàn một phần'],
    ]) {
      await deposit.locator('[name="kind"]').selectOption(kind);
      await deposit.locator('[name="room"]').selectOption(roomValue);
      await fields(deposit, {date:'2026-11-26', amount, reference:receipt, note});
      await deposit.getByRole('button', {name:'Lưu giao dịch cọc'}).click();
      await owner.getByRole('status').filter({hasText:'Đã ghi giao dịch'}).waitFor();
    }
    const metrics = owner.locator('.finance-metrics');
    const metric = label => metrics.locator('div').filter({hasText:label}).locator('dd');
    assert.equal(await metric('Chi phí vận hành').textContent(), '120.000 ₫');
    assert.equal(await metric('Cọc đã nhận').textContent(), '1.000.000 ₫');
    assert.equal(await metric('Cọc đã hoàn').textContent(), '400.000 ₫');
    assert.equal(await metric('Cọc đang giữ cuối kỳ').textContent(), '600.000 ₫');
    const statement = await owner.request.get(`${app}/cash.csv?period=${period}`);
    assert.equal(statement.status(), 200);
    const contents = await statement.text();
    assert.ok(contents.includes('Nhận cọc,' + roomName + ',,PW-COC,Cọc đầu kỳ,1000000,'));
    assert.ok(contents.includes('Hoàn cọc,' + roomName + ',,PW-HOAN,Hoàn một phần,,400000,'));
    assert.ok(contents.includes('Chi phí,,Bảo trì và sửa chữa,PW-CHI,Thay bóng đèn,,120000,'));
    assert.equal((await tenant.request.get(`${app}/cash.csv?period=${period}`, {maxRedirects:0})).status(), 303);
    const costRow = owner.locator('article').filter({hasText:'PW-CHI'});
    await costRow.getByText('Đính kèm chứng từ', {exact:true}).click();
    const pdf = Buffer.from('%PDF-1.4\n% ZentraSpace Playwright receipt\n%%EOF\n');
    await costRow.locator('input[type="file"]').setInputFiles({name:'receipt.pdf', mimeType:'application/pdf', buffer:pdf});
    await costRow.getByRole('button', {name:'Lưu chứng từ'}).click();
    await owner.getByRole('status').filter({hasText:'Đã đính kèm chứng từ'}).waitFor();
    const receiptURL = await costRow.getByRole('link', {name:'Tải chứng từ'}).getAttribute('href');
    const receipt = await owner.request.get(app + receiptURL);
    assert.equal(receipt.status(), 200);
    assert.ok((await receipt.body()).equals(pdf));
    assert.equal((await tenant.request.get(app + receiptURL, {maxRedirects:0})).status(), 303);
    await deposit.locator('[name="kind"]').selectOption('deposit_out');
    await deposit.locator('[name="room"]').selectOption(roomValue);
    await fields(deposit, {date:'2026-11-26', amount:'600001', note:'Kiểm tra giới hạn hoàn'});
    const invalidRefund = owner.waitForResponse(r => r.request().method() === 'POST' && r.url().endsWith('/action/cash-entry'));
    await deposit.getByRole('button', {name:'Lưu giao dịch cọc'}).click();
    assert.equal((await invalidRefund).status(), 422);
    console.log('PASS Playwright: expense, deposit, refund, balances, private receipt and CSV, refund guard');
    await checkRoomOnboarding(browser, owner, tenant);
    await checkPropertyWorkspace(browser, owner, tenant);
  } finally {
    await browser.close();
  }
}

run().catch(error => {
  console.error(error.message);
  process.exitCode = 1;
});
