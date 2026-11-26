const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { chromium } = require('playwright');

const app = process.env.ZS_BROWSER_BASE || 'http://localhost:8092';
const bank = process.env.ZS_BROWSER_MOCK || 'http://localhost:8093';
const envFile = process.env.ZS_BROWSER_ENV_FILE || path.resolve(__dirname, '../../.env');
const password = fs.readFileSync(envFile, 'utf8').match(/^ZS_OWNER_PASSWORD=([^\r\n]+)/m)?.[1];
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

async function run() {
  assert.ok(password, 'The test environment requires an owner password.');
  const channel = process.env.ZS_BROWSER_CHANNEL;
  const browser = await chromium.launch(channel ? {channel} : {});
  try {
    const owner = await browser.newPage();
    await owner.goto(app + '/login');
    await owner.getByLabel('Mật khẩu').fill(password);
    await owner.getByRole('button', {name:'Đăng nhập', exact:true}).click();
    await show(owner, 'Tổng quan kỳ thu');
    await owner.goto(`${app}/?tab=rooms&period=${period}`);
    await owner.getByRole('link', {name:'Thêm phòng', exact:true}).click();
    await fields(owner.locator('form[action="/action/room"]'), {
      name:roomName, tenant:'Người thuê mô phỏng', rent:'2000000',
      electricRate:'3000', waterRate:'10000', electric:'0', water:'0',
    });
    await owner.getByRole('button', {name:'Lưu phòng', exact:true}).click();
    await owner.getByRole('link', {name:`Xem chi tiết phòng ${roomName}`, exact:true}).click();
    await show(owner, roomName);
    const roomURL = owner.url();
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
  } finally {
    await browser.close();
  }
}

run().catch(error => {
  console.error(error.message);
  process.exitCode = 1;
});
